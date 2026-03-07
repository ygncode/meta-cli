# Auto-Reply Guide

Automatic Messenger replies powered by OpenClaw AI agents.

## Overview

When auto-reply is enabled, the webhook server automatically responds to incoming Messenger messages using an AI agent. The agent searches your knowledge base, reads conversation history, and sends helpful replies.

```
User sends FB message(s)
  → FB Webhook delivers to meta-cli server
  → meta-cli stores message(s) in SQLite
  → Debouncer collects messages per PSID (configurable, default 3s)
  → Timer fires (no new messages within window)
  → meta-cli POSTs to OpenClaw /hooks/agent
  → OpenClaw spins up isolated agent turn
  → Agent loads meta-cli-fb skill
  → Agent runs CLI commands:
      meta-cli rag search "user's question"
      meta-cli messenger history --psid USER
      meta-cli messenger send --psid USER -m "..."
  → Reply delivered through Messenger API
```

## Prerequisites

1. **meta-cli** installed and authenticated (`meta-cli auth login`)
2. **OpenClaw** installed and running with hooks enabled
3. A **knowledge base** directory with documents (Markdown, text, etc.)
4. A **Facebook Page** with webhook configured

## Setup Guide

### 1. Configure meta-cli

```bash
# Set your default page
meta-cli pages set-default YOUR_PAGE_ID

# Enable auto-reply
meta-cli config set auto_reply true

# Point to OpenClaw hooks endpoint
meta-cli config set hooks_endpoint http://127.0.0.1:18789/hooks/agent
meta-cli config set hooks_token YOUR_OPENCLAW_HOOKS_TOKEN

# Set knowledge base directory
meta-cli config set rag_dir ./knowledge-base

# Optional: adjust debounce window (default: 3 seconds)
meta-cli config set debounce_seconds 3
```

### 2. Install the OpenClaw Skill

```bash
# Copy skill to OpenClaw workspace
make install-skill

# Or manually:
cp -r skill/meta-cli-fb ~/.openclaw/workspace/skills/
```

### 3. Configure OpenClaw Hooks

Ensure OpenClaw has hooks enabled in `~/.openclaw/openclaw.json`:

```json
{
  "hooks": {
    "enabled": true,
    "token": "YOUR_OPENCLAW_HOOKS_TOKEN"
  }
}
```

The token must match what you set in `meta-cli config set hooks_token`.

### 4. Start the Webhook Server

```bash
# Foreground mode
meta-cli webhook serve --auto-reply

# Or as a daemon
meta-cli webhook serve --auto-reply --daemon
```

### 5. Test

Send a message to your Facebook Page via Messenger. You should see:
1. The message stored in the local database
2. After the debounce window, a POST to OpenClaw
3. The agent searches the knowledge base and sends a reply

## Configuration Reference

| Config Key | CLI Flag | Default | Description |
|-----------|----------|---------|-------------|
| `auto_reply` | `--auto-reply` | `false` | Enable the auto-reply pipeline |
| `hooks_endpoint` | `--hooks-endpoint` | `""` | OpenClaw `/hooks/agent` URL |
| `hooks_token` | `--hooks-token` | `""` | Bearer token for OpenClaw |
| `debounce_seconds` | `--debounce` | `3` | Seconds to wait before batching |
| `prompt_template` | (config only) | (built-in) | Go template for agent prompt |
| `rag_dir` | — | `""` | Knowledge base directory |

CLI flags override config values when both are provided.

## Prompt Template

The prompt template uses Go's `text/template` syntax. Available variables:

| Variable | Description |
|----------|-------------|
| `{{.PSID}}` | Page-scoped user ID |
| `{{.PageID}}` | Facebook Page ID |
| `{{.Messages}}` | List of batched messages |
| `{{range .Messages}}{{.Text}}{{end}}` | Iterate over message texts |

### Default Template

```
New message(s) from Facebook Messenger user (PSID: {{.PSID}}) on page {{.PageID}}:

{{range .Messages}}- {{.Text}}
{{end}}
Use the meta-cli skill to help this user. Search the knowledge base if needed, check conversation history for context, and send a helpful reply.
```

### Custom Template Example

```bash
meta-cli config set prompt_template "Customer {{.PSID}} asks:{{range .Messages}} {{.Text}}{{end}} - Reply in Burmese."
```

## Debouncing

When a user sends multiple messages quickly, the debouncer batches them into a single agent call:

```
msg1 at T=0  → start 3s timer
msg2 at T=1  → reset timer to 3s
msg3 at T=2  → reset timer to 3s
T=5          → fire callback with [msg1, msg2, msg3]
```

- Each user (PSID) has an independent timer
- Messages are delivered in arrival order
- The default window (3s) works well for most conversations
- Increase for users who type slowly; decrease for faster responses

## OpenClaw Skill

The `meta-cli-fb` skill teaches the OpenClaw agent to:

1. **Search the knowledge base** — `meta-cli rag search "<query>"`
2. **Read conversation history** — `meta-cli messenger history --psid <PSID>`
3. **Send replies** — `meta-cli messenger send --psid <PSID> --message "<reply>"`

### Rules

- Always search the knowledge base before replying
- Keep replies concise, friendly, and helpful
- If no relevant info is found, suggest contacting a human
- Use conversation history for context
- Never fabricate information

## Troubleshooting

### "hooks/agent error" in logs
- Check that OpenClaw is running (`openclaw status`)
- Verify `hooks_token` matches the token in OpenClaw's config
- Check `hooks_endpoint` URL is correct

### No replies being sent
- Verify the `meta-cli-fb` skill is installed in OpenClaw workspace
- Check `rag_dir` is set and contains documents
- Check OpenClaw logs for agent errors

### Replies are slow
- Reduce `debounce_seconds` for faster response
- Check the model speed in OpenClaw configuration
- Consider using a faster model

### Duplicate replies
- Message deduplication is handled by SQLite's `INSERT OR IGNORE`
- Echo messages (outgoing) are stored but not debounced

### Agent can't find meta-cli
- Ensure `meta-cli` is in the system PATH
- The skill metadata requires `meta-cli` binary: `{"openclaw": {"requires": {"bins": ["meta-cli"]}}}`
