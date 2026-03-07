# Webhook System

## Overview

The webhook system enables real-time message reception from Facebook Messenger. When someone sends a message to your Facebook Page, Meta delivers the event to your webhook server, which stores the message locally in SQLite.

## Architecture

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
```

## Components

### WebhookHandler

**File:** `internal/messenger/webhook.go`

The `WebhookHandler` implements `http.Handler` and processes two types of HTTP requests:

```go
type WebhookHandler struct {
    VerifyToken string              // Token for Meta's verification handshake
    AppSecret   string              // App secret for HMAC signature validation
    PageID      string              // Page ID for message association
    Store       *Store              // SQLite message store
    Messenger   *Service            // Messenger service
    Debouncer   DebouncerInterface  // Per-PSID debouncer (nil = auto-reply disabled)
}
```

### Verification (GET)

When Meta sets up the webhook, it sends a GET request to verify ownership:

```
GET /webhook?hub.mode=subscribe&hub.verify_token=YOUR_TOKEN&hub.challenge=CHALLENGE
```

The handler:
1. Checks that `hub.mode` is `"subscribe"`
2. Validates `hub.verify_token` matches the configured token
3. Returns the `hub.challenge` value with HTTP 200
4. Returns HTTP 403 if verification fails

### Event Reception (POST)

When a message is sent to or from the page, Meta delivers a POST request:

```
POST /webhook
X-Hub-Signature-256: sha256=HMAC_HASH
Content-Type: application/json

{
  "object": "page",
  "entry": [{
    "id": "PAGE_ID",
    "time": 1234567890,
    "messaging": [{
      "sender": {"id": "USER_PSID"},
      "recipient": {"id": "PAGE_ID"},
      "timestamp": 1234567890000,
      "message": {
        "mid": "m_MESSAGE_ID",
        "text": "Hello!",
        "is_echo": false
      }
    }]
  }]
}
```

The handler:
1. Reads the request body
2. Validates the HMAC-SHA256 signature using the app secret
3. Immediately responds with HTTP 200 and `"EVENT_RECEIVED"`
4. Processes the payload asynchronously in a goroutine

### Signature Validation

```go
func (h *WebhookHandler) validateSignature(body []byte, signature string) bool {
    // 1. Require app secret to be set
    // 2. Check signature starts with "sha256="
    // 3. Compute HMAC-SHA256 of body using app secret
    // 4. Compare with constant-time comparison (hmac.Equal)
}
```

The `X-Hub-Signature-256` header contains `sha256=` followed by the hex-encoded HMAC-SHA256 hash of the request body, computed using the app secret as the key.

### Message Processing

For each messaging event in the payload:

1. **Skip non-text messages** - Events without a message or with empty text are ignored
2. **Determine direction:**
   - `is_echo: true` → Outgoing message (sent by the page). PSID = `recipient.id`, direction = `"out"`
   - `is_echo: false` → Incoming message (sent by user). PSID = `sender.id`, direction = `"in"`
3. **Deduplicate** - Check if the message ID already exists in the database
4. **Store** - Save the message to SQLite with all metadata
5. **Log** - Print the message to the server log

## Webhook Subscription

Before the server can receive events, the page must be subscribed to webhook fields.

**File:** `internal/messenger/service.go`

```go
func (s *Service) SubscribeWebhook(ctx context.Context) error {
    // POST /me/subscribed_apps
    // Body: subscribed_fields=messages,message_echoes
}
```

Subscribed fields:
- `messages` - Incoming messages from users
- `message_echoes` - Outgoing messages sent by the page (including via API)

The `webhook serve` command automatically calls `SubscribeWebhook` before starting the HTTP server.

## Daemon Mode

**File:** `internal/daemon/daemon.go`

When `--daemon` is passed, the webhook server runs as a background process:

### Starting the Daemon

```bash
meta-cli webhook serve --verify-token MY_TOKEN --daemon
```

1. Forks the current process using `os.StartProcess`
2. Redirects stdout/stderr to `~/.meta-cli/webhook.log`
3. Writes the child PID to `~/.meta-cli/webhook.pid`
4. Parent process exits, child continues running

### Daemon File Locations

| File | Path | Purpose |
|------|------|---------|
| PID file | `~/.meta-cli/webhook.pid` | Store daemon process ID |
| Log file | `~/.meta-cli/webhook.log` | Stdout/stderr output |

### Checking Status

```bash
meta-cli webhook status
```

1. Reads PID from `~/.meta-cli/webhook.pid`
2. Checks if process is running with `syscall.Kill(pid, 0)`
3. Reports status and PID

### Stopping the Daemon

```bash
meta-cli webhook stop
```

1. Reads PID from the PID file
2. Sends `SIGTERM` to the process
3. Waits up to 3 seconds for graceful shutdown
4. Removes the PID file

### Graceful Shutdown

The server handles `SIGINT` and `SIGTERM` signals:

```
Signal received
  └── HTTP server.Shutdown(ctx)  # Stop accepting new connections
       └── Wait for active requests to complete
            └── Close SQLite database
                 └── Exit
```

## Setup Guide

To receive webhooks, you need:

1. **A publicly accessible URL** - Use a tunnel service (e.g., ngrok) during development
2. **Configure webhook in Meta App Dashboard:**
   - Go to your app dashboard → Webhooks
   - Subscribe to "Page" object
   - Set callback URL to your public URL + `/webhook`
   - Set verify token to match your `--verify-token` value
3. **Start the webhook server:**
   ```bash
   meta-cli webhook serve --verify-token YOUR_TOKEN --port 8080
   ```
4. **In another terminal, expose the port:**
   ```bash
   ngrok http 8080
   ```

## Data Flow Summary

```
User sends message to Page on Messenger
  │
  ▼
Meta Platform detects new message
  │
  ▼
POST webhook event to your server
  │
  ▼
WebhookHandler.receive()
  ├── Validate HMAC-SHA256 signature
  ├── Respond 200 immediately
  └── Async: processPayload()
        ├── Parse JSON payload
        ├── For each messaging event:
        │     ├── Determine direction (in/out based on is_echo)
        │     ├── Check deduplication (MessageExists)
        │     └── SaveMessage to SQLite
        └── Log message details

Page sends message via API (messenger send)
  │
  ▼
Meta Platform creates message_echo event
  │
  ▼
Same webhook flow, but direction = "out"
```

## Auto-Reply Pipeline

When `auto_reply` is enabled, the webhook server integrates with [OpenClaw](https://github.com/openclaw) to automatically respond to incoming messages using an AI agent.

### Architecture

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

### Message Debouncing

When a user sends multiple messages rapidly, the debouncer collects them before triggering the agent:

```
msg1 from user_1 at T=0  → start timer(user_1, 3s)
msg2 from user_1 at T=1  → reset timer(user_1, 3s)
msg3 from user_1 at T=2  → reset timer(user_1, 3s)
T=5 (3s after last msg)  → callback(user_1, [msg1, msg2, msg3])
```

Each PSID has an independent timer. The debounce window is configurable via `debounce_seconds` (default: 3).

### OpenClaw Integration

After debouncing, a POST request is sent to OpenClaw's `/hooks/agent` endpoint:

```json
{
    "message": "<rendered prompt with batched messages>",
    "name": "FB Messenger",
    "deliver": false,
    "sessionKey": "hook:fb:<psid>"
}
```

- `deliver: false` — the agent sends replies itself via `meta-cli messenger send`
- `sessionKey` — each FB user gets their own isolated session for cross-turn context

### Configuration

| Config Key | Default | Description |
|-----------|---------|-------------|
| `auto_reply` | `false` | Enable auto-reply pipeline |
| `hooks_endpoint` | `""` | OpenClaw `/hooks/agent` URL |
| `hooks_token` | `""` | Bearer token for OpenClaw auth |
| `debounce_seconds` | `3` | Seconds to wait before batching |
| `prompt_template` | (built-in) | Go template for the agent prompt |

### Troubleshooting

- **"hooks/agent error" in logs** — check OpenClaw is running and token matches
- **No replies being sent** — verify the `meta-cli-fb` skill is installed in OpenClaw and `rag_dir` is set
- **Replies are slow** — check `debounce_seconds` and the model speed in OpenClaw
- **Duplicate replies** — message deduplication is handled by the store's `INSERT OR IGNORE`
