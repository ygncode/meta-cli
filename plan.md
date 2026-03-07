# Auto-Reply Pipeline Plan

## Goal

When a user sends a message to a Facebook Page via Messenger, the CLI webhook receives it, debounces messages, then calls **OpenClaw** via its `/hooks/agent` endpoint. OpenClaw runs an isolated agent turn with a **meta-cli skill** that teaches it to use `meta-cli` commands (RAG search, conversation history, send reply). The agent decides what to do autonomously.

## How OpenClaw `/hooks/agent` Works

Based on source code review (`openclaw/src/gateway/hooks.ts`, `server/hooks.ts`, `cron/isolated-agent/run.ts`):

1. **POST `/hooks/agent`** accepts JSON with `message` (required), `name`, `deliver`, `sessionKey`, `model`, `thinking`, `timeoutSeconds`
2. Validates auth via `Authorization: Bearer <token>` or `x-openclaw-token` header
3. Returns `200 { ok: true, runId: "..." }` immediately (async)
4. Internally creates an **isolated agent turn** (via `runCronIsolatedAgentTurn`)
5. The agent run uses **pi (coding agent)** as the runtime — it has `exec`/`bash` tool access
6. The agent sees workspace skills in its system prompt and can execute CLI commands
7. After the run completes, a summary is posted to the main session

**Key:** `deliver: false` since the agent will send replies itself via `meta-cli messenger send`. We don't want OpenClaw to also deliver to a chat channel.

## Flow

```
User sends FB message(s)
  → FB Webhook delivers to meta-cli server
  → meta-cli stores message(s) in SQLite
  → Debouncer collects messages per PSID (configurable, default 3s)
  → Timer fires (no new messages within window)
  → meta-cli POSTs to OpenClaw /hooks/agent:
      {
        "message": "<rendered prompt with batched messages>",
        "name": "FB Messenger",
        "deliver": false,
        "sessionKey": "hook:fb:<psid>"
      }
  → OpenClaw spins up isolated agent turn
  → Agent loads meta-cli-fb skill
  → Agent runs CLI commands:
      meta-cli rag search "user's question"       → knowledge base
      meta-cli messenger history --psid USER       → conversation context
      meta-cli messenger send --psid USER -m "..." → send reply
  → Reply delivered through Messenger API
```

## Current State

### Already Have ✅
- Webhook server (FB message receive, HMAC validation, SQLite storage)
- Message store (save/list/exists/markAutoReplied)
- RAG engine (TF-IDF search via `meta-cli rag search`)
- Messenger send (`meta-cli messenger send`)
- Auth/config/pages/daemon — all working

### Need to Build 🔴
1. `messenger history` CLI command (so agent can read conversation context)
2. Message debouncer (per-PSID timer, configurable window)
3. OpenClaw hooks caller (POST to `/hooks/agent` after debounce)
4. Config fields (debounce_seconds, hooks_endpoint, hooks_token, auto_reply, prompt_template)
5. Wire debouncer + hooks into webhook handler
6. OpenClaw skill (`SKILL.md` teaching the agent meta-cli commands)

---

## Phase 1: Config Changes

**File:** `internal/config/types.go`

```go
type Config struct {
    // ... existing fields ...
    DebounceSeconds int    `json:"debounce_seconds,omitempty"`   // default: 3
    HooksEndpoint   string `json:"hooks_endpoint,omitempty"`     // e.g. "http://127.0.0.1:18789/hooks/agent"
    HooksToken      string `json:"hooks_token,omitempty"`        // OpenClaw hooks auth token
    AutoReply       bool   `json:"auto_reply,omitempty"`         // enable/disable auto-reply
    PromptTemplate  string `json:"prompt_template,omitempty"`    // Go template for the agent prompt
}
```

**File:** `internal/config/config.go` — update `Default()` to set `DebounceSeconds: 3`.

**File:** `cmd_impl/config.go` — add all 5 new keys to `setConfigField`, `getConfigField`, and `configListCmd`.

**Default prompt template:**
```
New message(s) from Facebook Messenger user (PSID: {{.PSID}}) on page {{.PageID}}:

{{range .Messages}}- {{.Text}}
{{end}}

Use the meta-cli skill to help this user. Search the knowledge base if needed, check conversation history for context, and send a helpful reply.
```

The prompt template is configurable via `meta-cli config set prompt_template "..."` so users can customize agent behavior per their use case.

### Phase 1 Tests

**File:** `internal/config/config_test.go` — add:
- `TestDefaultDebounceSeconds` — verify default is 3
- `TestSaveAndLoadAutoReply` — round-trip new fields through save/load
- `TestSetConfigFieldNewKeys` — verify all 5 new keys work in `setConfigField`
- `TestGetConfigFieldNewKeys` — verify all 5 new keys work in `getConfigField`

---

## Phase 2: Messenger History Command

**File:** `internal/messenger/store.go`

Add method:
```go
func (s *Store) RecentMessages(pageID, psid string, limit int) ([]Message, error)
```

Returns last N messages (both in/out) for a PSID, ordered chronologically (oldest first).

**File:** `cmd_impl/messenger.go`

Add `messenger history` subcommand:
```bash
meta-cli messenger history --psid USER_PSID [--limit 20] [--json]
```

Output format (table):
```
DIRECTION  TEXT                    RECEIVED_AT
in         How do I reset my pw?   2026-03-07T09:00:00Z
out        Go to Settings > ...    2026-03-07T09:00:05Z
in         Thanks!                 2026-03-07T09:01:00Z
```

This is what OpenClaw's agent will call to get conversation context.

### Phase 2 Tests

**File:** `internal/messenger/store_test.go` — add:
- `TestRecentMessages` — insert messages for multiple PSIDs, verify filtering by PSID + page
- `TestRecentMessagesOrder` — verify chronological order (oldest first)
- `TestRecentMessagesLimit` — verify limit is respected
- `TestRecentMessagesBothDirections` — insert in + out messages, verify both returned
- `TestRecentMessagesEmpty` — verify empty result for unknown PSID

---

## Phase 3: Message Debouncer

**New package:** `internal/debounce/`

```
internal/debounce/
  debounce.go
  debounce_test.go
```

Interface:
```go
type Message struct {
    ID   string
    Text string
}

type Callback func(psid string, messages []Message)

type Debouncer struct { ... }

func New(window time.Duration, cb Callback) *Debouncer
func (d *Debouncer) Add(psid string, msg Message)
func (d *Debouncer) Stop()  // clean shutdown, cancel all pending timers
```

Behavior:
```
msg1 from user_1 at T=0  → start timer(user_1, 3s)
msg2 from user_1 at T=1  → reset timer(user_1, 3s)
msg3 from user_1 at T=2  → reset timer(user_1, 3s)
T=5 (3s after last msg)  → callback(user_1, [msg1, msg2, msg3])
```

Thread-safe. Per-PSID independent timers. Uses `sync.Mutex` + `time.AfterFunc`.

### Phase 3 Tests

**File:** `internal/debounce/debounce_test.go`:
- `TestSingleMessage` — one message, verify callback fires after window
- `TestDebounceResets` — multiple messages within window, verify callback fires once with all messages
- `TestMultiplePSIDs` — two users sending concurrently, verify independent timers and separate callbacks
- `TestMessageOrder` — verify messages are delivered in arrival order
- `TestStop` — call Stop(), verify pending timers are cancelled and no callback fires
- `TestStopIdempotent` — calling Stop() twice doesn't panic
- `TestAddAfterStop` — adding after Stop() doesn't panic (no-op or graceful)
- `TestZeroWindow` — zero duration fires callback immediately
- `TestConcurrentAdds` — multiple goroutines adding to same PSID simultaneously (race detector)

---

## Phase 4: OpenClaw Hooks Caller

**New package:** `internal/hooks/`

```
internal/hooks/
  hooks.go
  hooks_test.go
```

Interface:
```go
type Client struct {
    endpoint   string
    token      string
    httpClient *http.Client
}

func NewClient(endpoint, token string) *Client
func NewClientWithHTTP(endpoint, token string, hc *http.Client) *Client  // for testing
func (c *Client) CallAgent(ctx context.Context, prompt string, psid string) error
```

The POST body follows OpenClaw's exact `/hooks/agent` contract:
```json
{
    "message": "<rendered prompt>",
    "name": "FB Messenger",
    "deliver": false,
    "sessionKey": "hook:fb:<psid>"
}
```

- `deliver: false` → agent sends reply itself via `meta-cli messenger send`
- `sessionKey: "hook:fb:<psid>"` → each FB user gets their own isolated session, so agent remembers per-user context across turns
- Response: `200 { ok: true, runId: "..." }` — fire and forget, async

Also add prompt template rendering:
```go
func RenderPrompt(tmpl string, psid, pageID string, messages []debounce.Message) (string, error)
```

### Phase 4 Tests

**File:** `internal/hooks/hooks_test.go`:
- `TestCallAgentSuccess` — httptest server returns 200 `{"ok":true,"runId":"xxx"}`, verify no error
- `TestCallAgentRequestFormat` — verify POST body contains correct JSON fields (message, name, deliver=false, sessionKey)
- `TestCallAgentAuthHeader` — verify `Authorization: Bearer <token>` header is sent
- `TestCallAgentServerError` — server returns 500, verify error returned
- `TestCallAgentUnauthorized` — server returns 401, verify error returned
- `TestCallAgentTimeout` — server hangs, cancelled context, verify error
- `TestCallAgentBadEndpoint` — invalid URL, verify error
- `TestRenderPrompt` — verify template rendering with PSID, PageID, multiple messages
- `TestRenderPromptDefault` — verify default template renders correctly
- `TestRenderPromptInvalid` — invalid Go template, verify error

---

## Phase 5: Wire Into Webhook Handler

**File:** `internal/messenger/webhook.go`

Add to `WebhookHandler`:
```go
type WebhookHandler struct {
    // ... existing fields ...
    Debouncer  *debounce.Debouncer  // NEW (nil = auto-reply disabled)
}
```

Change `processPayload`:
1. Parse message → store in DB (same as now)
2. If message is inbound (not echo) AND `Debouncer != nil`:
   - `Debouncer.Add(senderPSID, msg)`
3. Everything else unchanged

**File:** `cmd_impl/webhook.go`

Update `webhook serve`:
- If `auto_reply` is true (config or `--auto-reply` flag):
  - Validate `hooks_endpoint` and `hooks_token` are set
  - Initialize hooks client
  - Build prompt renderer from `prompt_template`
  - Initialize debouncer with callback:
    ```go
    func(psid string, msgs []debounce.Message) {
        prompt := renderPrompt(template, psid, pageID, msgs)
        if err := hooksClient.CallAgent(ctx, prompt, psid); err != nil {
            log.Printf("hooks/agent error for %s: %v", psid, err)
        }
    }
    ```
  - Pass debouncer into WebhookHandler
  - Defer `debouncer.Stop()` for graceful shutdown

New flags:
```
--auto-reply              Enable auto-reply (overrides config)
--debounce <seconds>      Debounce window in seconds (overrides config)
--hooks-endpoint <url>    OpenClaw hooks endpoint (overrides config)
--hooks-token <token>     OpenClaw hooks token (overrides config)
```

### Phase 5 Tests

**File:** `internal/messenger/webhook_test.go` — add:
- `TestWebhookWithDebouncer` — send valid webhook POST with debouncer set, verify message is fed to debouncer
- `TestWebhookWithoutDebouncer` — Debouncer is nil, verify messages still stored but no debounce called (backward compatible)
- `TestWebhookEchoNotDebounced` — echo messages should NOT be fed to debouncer
- `TestWebhookAutoReplyMissingConfig` — verify error when auto_reply=true but hooks_endpoint/hooks_token missing (test in cmd_impl level)

---

## Phase 6: OpenClaw Skill

**File:** `skill/meta-cli-fb/SKILL.md`

```markdown
---
name: meta-cli-fb
description: >
  Manage Facebook Page Messenger conversations using meta-cli.
  Search knowledge base via RAG, read conversation history,
  and send replies. Use when handling Facebook Messenger webhook messages.
metadata: {"openclaw": {"requires": {"bins": ["meta-cli"]}}}
---

# meta-cli Facebook Messenger Skill

You manage a Facebook Page's Messenger inbox. When triggered by a
webhook, you receive batched user messages and respond using the
page's knowledge base.

## Commands

### Search Knowledge Base
\`\`\`bash
meta-cli rag search "<query>" [--dir <path>] [--top 5] [--json]
\`\`\`
Search documents for relevant answers. Always do this before replying.

### Read Conversation History
\`\`\`bash
meta-cli messenger history --psid <PSID> [--limit 20] [--json]
\`\`\`
Get recent messages (both directions) with this user.

### Send Reply
\`\`\`bash
meta-cli messenger send --psid <PSID> --message "<reply>"
\`\`\`
Send a Messenger reply to the user.

### List Recent Messages
\`\`\`bash
meta-cli messenger list [--limit 50] [--json]
\`\`\`

## Workflow

1. Parse the PSID and message(s) from the prompt
2. Search knowledge base: `meta-cli rag search "<user question>"`
3. If needed, check history: `meta-cli messenger history --psid <PSID>`
4. Compose a reply based on knowledge base results
5. Send: `meta-cli messenger send --psid <PSID> --message "<reply>"`

## Rules

- Always search the knowledge base before replying
- Keep replies concise, friendly, and helpful
- If the knowledge base has no relevant info, say so honestly
  and suggest the user contact a human
- Use conversation history to maintain context
- Do NOT make up information not found in the knowledge base
```

### Installation
```bash
# Copy to OpenClaw workspace
cp -r skill/meta-cli-fb ~/.openclaw/workspace/skills/

# Or link it
ln -s $(pwd)/skill/meta-cli-fb ~/.openclaw/workspace/skills/meta-cli-fb
```

---

## Phase 7: Documentation Updates

### 7.1 Update `README.md`

Add to the Quick Start section:
```bash
# --- Auto-Reply (with OpenClaw) ---
meta-cli config set auto_reply true
meta-cli config set hooks_endpoint http://127.0.0.1:18789/hooks/agent
meta-cli config set hooks_token YOUR_TOKEN
meta-cli config set rag_dir ./knowledge-base
meta-cli webhook serve --auto-reply --daemon
```

Add to Commands table:
| `messenger history` | List conversation history with a user |

Add to Config section — document new config fields:
| `debounce_seconds` | Debounce window for message batching (default: 3) |
| `hooks_endpoint` | OpenClaw webhook endpoint URL |
| `hooks_token` | OpenClaw webhook auth token |
| `auto_reply` | Enable auto-reply pipeline (true/false) |
| `prompt_template` | Go template for agent prompt |

Add new Architecture section entry:
```
internal/
  debounce/              # Message debouncing (per-user timer)
  hooks/                 # OpenClaw /hooks/agent caller
```

Update webhook serve flags table to include new flags.

### 7.2 Update `docs/commands.md`

Add under **Messenger Commands**:

#### `messenger history`

Full documentation matching the existing command doc style:
- Usage, flags (`--psid`, `--limit`), output format
- Note that this reads from local DB
- Add to Command Requirements Matrix (requires Page ID, no auth)

Add under **Webhook Commands** > `webhook serve`:
- Document new flags: `--auto-reply`, `--debounce`, `--hooks-endpoint`, `--hooks-token`
- Document auto-reply behavior

Add new config keys to **Config Commands** > `config set` supported keys list.

### 7.3 Update `docs/webhooks.md`

Add new section **## Auto-Reply Pipeline** after the existing Data Flow Summary:

Document:
- Overview of the auto-reply flow
- Message debouncing explanation with timing diagram
- OpenClaw integration (how `/hooks/agent` is called)
- Prompt template system
- Configuration reference
- Troubleshooting (OpenClaw not running, missing config, etc.)

Update the architecture diagram:
```
Facebook Messenger
       │
       ▼
Meta Platform ──── POST /webhook ────► meta-cli webhook server
                                              │
                                    ┌─────────┼─────────┐
                                    ▼         ▼         ▼
                              Signature    Parse     Store in
                              Validation   Payload   SQLite DB
                                                        │
                                                        ▼
                                                   Debouncer
                                                   (per-PSID)
                                                        │
                                                        ▼ (after quiet period)
                                                   POST /hooks/agent
                                                   (OpenClaw)
                                                        │
                                                        ▼
                                                   Agent runs:
                                                   - rag search
                                                   - messenger history
                                                   - messenger send
```

Update the `WebhookHandler` struct documentation to include the `Debouncer` field.

### 7.4 Update `docs/architecture.md`

Add to Package Dependency Graph:
```
cmd_impl
  ├── internal/debounce    (message debouncing)
  ├── internal/hooks       (OpenClaw hook caller)
  ... (existing)
```

Add to the Architecture diagram:
```
internal/
  ├── debounce/            # Per-PSID message debouncing
  ├── hooks/               # OpenClaw /hooks/agent integration
  ... (existing)
```

Update the Execution Flow for `webhook serve` to show the auto-reply initialization path.

### 7.5 Update `docs/storage.md`

Add new config fields to the Config Fields table:

| `debounce_seconds` | int | `3` | Seconds to wait before batching messages |
| `hooks_endpoint` | string | `""` | OpenClaw /hooks/agent endpoint URL |
| `hooks_token` | string | `""` | Bearer token for OpenClaw hooks auth |
| `auto_reply` | bool | `false` | Enable auto-reply via OpenClaw |
| `prompt_template` | string | `""` | Go template for agent prompts |

Update the Example Config JSON to include new fields.

Update the Config Structure go code block.

### 7.6 Update `docs/development.md`

Add to Project Structure:
```
├── internal/
│   ├── debounce/             # Message debouncing
│   │   ├── debounce.go       # Per-PSID timer with batching
│   │   └── debounce_test.go
│   ├── hooks/                # OpenClaw integration
│   │   ├── hooks.go          # /hooks/agent caller + prompt rendering
│   │   └── hooks_test.go
```

Add to the "Adding a New Command" section — mention `messenger history` as an example of a command that reads local DB without API access.

### 7.7 New doc `docs/auto-reply.md`

Create a dedicated auto-reply guide covering:

1. **Overview** — what it does, the full flow diagram
2. **Prerequisites** — OpenClaw installed and running, hooks enabled, meta-cli authenticated
3. **Setup Guide** — step-by-step:
   - Configure meta-cli (config set commands)
   - Install the OpenClaw skill
   - Configure OpenClaw hooks
   - Start the webhook server
   - Test with a message
4. **Configuration Reference** — all config fields with descriptions
5. **Prompt Template** — Go template syntax, available variables (`.PSID`, `.PageID`, `.Messages`), examples
6. **Debouncing** — how it works, why 3s default, how to tune
7. **OpenClaw Skill** — what commands the agent uses, how to customize
8. **Troubleshooting** — common issues:
   - "hooks/agent error" in logs → check OpenClaw is running, token matches
   - No replies being sent → check skill is installed, rag_dir is set
   - Replies are slow → check debounce_seconds, OpenClaw model speed
   - Duplicate replies → check message deduplication in store

Update `docs/README.md` table of contents to include the new doc:
| [Auto-Reply Guide](./auto-reply.md) | OpenClaw integration for automatic Messenger replies |

### 7.8 Update `Makefile`

Add:
```makefile
install-skill:
	mkdir -p ~/.openclaw/workspace/skills/
	cp -r skill/meta-cli-fb ~/.openclaw/workspace/skills/
```

---

## Phase 8: Test Summary

### New Test Files

| File | Tests |
|------|-------|
| `internal/debounce/debounce_test.go` | 9 tests (single msg, debounce reset, multi-PSID, order, stop, stop idempotent, add after stop, zero window, concurrent adds) |
| `internal/hooks/hooks_test.go` | 10 tests (success, request format, auth header, server error, unauthorized, timeout, bad endpoint, render prompt, render default, render invalid) |

### Updated Test Files

| File | New Tests |
|------|-----------|
| `internal/config/config_test.go` | 4 tests (default debounce, save/load auto-reply fields, set new keys, get new keys) |
| `internal/messenger/store_test.go` | 5 tests (recent messages, order, limit, both directions, empty) |
| `internal/messenger/webhook_test.go` | 4 tests (with debouncer, without debouncer, echo not debounced, auto-reply missing config) |

### Test Patterns to Follow

Based on existing tests:
- Use `httptest.NewServer` for API mocking (see `graph_test.go`, `posts_test.go`)
- Use `":memory:"` SQLite for store tests (see `store_test.go`)
- Use `t.TempDir()` + `t.Setenv("HOME", tmp)` for config tests (see `config_test.go`)
- Use `graph.NewWithHTTPClient(srv.URL, token, srv.Client())` for graph client tests
- Keep tests independent — no shared state between test functions
- Test both success and error paths
- For the debouncer, use short windows (10ms-50ms) to keep tests fast

### Running Tests
```bash
# All tests
make test

# Specific package
go test ./internal/debounce/...
go test ./internal/hooks/...

# With race detector (important for debouncer)
go test -race ./internal/debounce/...

# Verbose
go test -v ./...
```

---

## File Change Summary

```
Modified:
  internal/config/types.go              → add 5 new config fields
  internal/config/config.go             → update Default()
  internal/config/config_test.go        → 4 new tests
  internal/messenger/webhook.go         → add Debouncer field, feed inbound msgs
  internal/messenger/webhook_test.go    → 4 new tests
  internal/messenger/store.go           → add RecentMessages()
  internal/messenger/store_test.go      → 5 new tests
  cmd_impl/config.go                    → add new config keys to set/get/list
  cmd_impl/webhook.go                   → wire debouncer + hooks, add flags
  cmd_impl/messenger.go                 → add messenger history command
  Makefile                              → add install-skill target
  README.md                             → document auto-reply setup + new commands
  docs/README.md                        → add auto-reply.md to table of contents
  docs/commands.md                      → add messenger history, new webhook flags, new config keys
  docs/webhooks.md                      → add Auto-Reply Pipeline section, update diagrams
  docs/architecture.md                  → add debounce + hooks packages
  docs/storage.md                       → add new config fields, update example
  docs/development.md                   → add new packages to structure

New:
  internal/debounce/debounce.go
  internal/debounce/debounce_test.go    → 9 tests
  internal/hooks/hooks.go
  internal/hooks/hooks_test.go          → 10 tests
  skill/meta-cli-fb/SKILL.md
  docs/auto-reply.md                    → dedicated auto-reply guide

No removals.
```

---

## Open Questions

### Q1: Prompt Template Customization
The prompt template is stored as a single string in config. Should we also support:
- **A)** Single string in config only (simplest — user does `meta-cli config set prompt_template "..."`)
- **B)** Also support a file path (e.g. `prompt_template_file` pointing to a .txt file for longer prompts)

### Q2: Auto-Reply Filtering
Should the pipeline skip certain inbound messages?
- **A)** No filtering — always trigger (simplest, let the agent decide)
- **B)** Skip non-text messages only (stickers, reactions, images with no text)
- **C)** Configurable skip patterns

### Q3: Failure Handling
When OpenClaw `/hooks/agent` returns non-200 or is unreachable:
- **A)** Log error, skip silently (messages stay in DB for manual handling)
- **B)** Configurable fallback message in config (empty = silent skip)
- **C)** Retry with backoff, then silent skip

### Q4: Session Key Strategy
Each `/hooks/agent` call uses `sessionKey: "hook:fb:<psid>"`. This means:
- Same user gets continuity across turns (agent remembers prior interactions)
- Different users are fully isolated
- **Is this the right default?** Or should all FB messages share one session?

### Q5: Skill Location
Where should the skill live?
- **A)** In this repo under `skill/` — user copies to OpenClaw workspace
- **B)** Auto-installed via `meta-cli config set auto_reply true` (copies skill automatically)
- **C)** Published to ClawHub for `clawhub install meta-cli-fb`

---

## Setup Flow (User Perspective)

```bash
# 1. Install & auth meta-cli (already works)
meta-cli auth login --app-id APP_ID --app-secret APP_SECRET
meta-cli pages set-default PAGE_ID

# 2. Configure auto-reply
meta-cli config set auto_reply true
meta-cli config set hooks_endpoint http://127.0.0.1:18789/hooks/agent
meta-cli config set hooks_token YOUR_OPENCLAW_HOOKS_TOKEN
meta-cli config set rag_dir ./knowledge-base
meta-cli config set debounce_seconds 3

# 3. Install the skill into OpenClaw
make install-skill
# or: cp -r skill/meta-cli-fb ~/.openclaw/workspace/skills/

# 4. Make sure OpenClaw has hooks enabled in ~/.openclaw/openclaw.json:
# { "hooks": { "enabled": true, "token": "YOUR_OPENCLAW_HOOKS_TOKEN" } }

# 5. Start webhook server
meta-cli webhook serve --auto-reply --daemon

# Done! Messages now flow:
# FB → webhook → debounce → OpenClaw → rag search → reply
```

---

## Implementation Order

When implementing, follow this order strictly. Each phase builds on the previous.

1. **Phase 1** — Config changes (foundation for everything)
2. **Phase 2** — Messenger history command (standalone, no dependencies on other new code)
3. **Phase 3** — Debouncer package (standalone, tested independently)
4. **Phase 4** — Hooks caller package (standalone, tested independently)
5. **Phase 5** — Wire into webhook (depends on Phase 1, 3, 4)
6. **Phase 6** — Skill file (standalone, no code changes)
7. **Phase 7** — Documentation (after all code is done)
8. **Phase 8** — Final test pass, run `make test` and `go test -race ./...`

After each phase, run `make test` to verify nothing is broken.
