# Data Storage

meta-cli uses three storage mechanisms, each suited to its purpose.

## Storage Overview

| Storage | Technology | Location | Purpose |
|---------|-----------|----------|---------|
| Configuration | JSON file | `~/.meta-cli/config.json` | App settings |
| Credentials | OS Keyring | System keychain | Tokens and secrets |
| Messages | SQLite | `~/.meta-cli/messages.db` | Message history |

## Configuration

**File:** `internal/config/config.go`, `internal/config/types.go`

### Config File Location

`~/.meta-cli/config.json`

The directory `~/.meta-cli/` is created automatically with `0700` permissions on first use.

### Config Structure

```go
type Config struct {
    DefaultAccount  string             `json:"default_account"`
    DefaultPage     string             `json:"default_page"`
    GraphAPIVersion string             `json:"graph_api_version"`
    WebhookPort     int                `json:"webhook_port"`
    RAGDir          string             `json:"rag_dir"`
    DBPath          string             `json:"db_path"`
    VerifyToken     string             `json:"verify_token"`
    Accounts        map[string]Account `json:"accounts,omitempty"`
    RedirectURI     string             `json:"redirect_uri"`

    // Auto-reply pipeline
    DebounceSeconds int    `json:"debounce_seconds"`
    HooksEndpoint   string `json:"hooks_endpoint"`
    HooksToken      string `json:"hooks_token"`
    AutoReply       bool   `json:"auto_reply"`
    PromptTemplate  string `json:"prompt_template"`
}

type Account struct {
    AppID string `json:"app_id"`
}
```

### Config Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default_account` | string | `"default"` | Account name used when `--account` is not specified |
| `default_page` | string | `""` | Page ID used when `--page` is not specified |
| `graph_api_version` | string | `"v25.0"` | Meta Graph API version |
| `webhook_port` | int | `8080` | Default port for the webhook server |
| `rag_dir` | string | `""` | Directory containing documents for RAG search |
| `db_path` | string | `""` | Custom SQLite database path (empty = default) |
| `verify_token` | string | `""` | Webhook verification token |
| `accounts` | map | `{}` | Per-account configurations with App IDs |
| `redirect_uri` | string | `""` | Custom OAuth redirect URI (default: `https://localhost/`) |
| `debounce_seconds` | int | `3` | Seconds to wait before batching messages |
| `hooks_endpoint` | string | `""` | OpenClaw `/hooks/agent` endpoint URL |
| `hooks_token` | string | `""` | Bearer token for OpenClaw hooks auth |
| `auto_reply` | bool | `false` | Enable auto-reply via OpenClaw |
| `prompt_template` | string | `""` | Go template for agent prompts |

### Example Config

```json
{
  "default_account": "default",
  "default_page": "123456789",
  "graph_api_version": "v25.0",
  "webhook_port": 8080,
  "rag_dir": "./docs",
  "db_path": "",
  "verify_token": "",
  "redirect_uri": "",
  "debounce_seconds": 3,
  "hooks_endpoint": "",
  "hooks_token": "",
  "auto_reply": false,
  "prompt_template": "",
  "accounts": {
    "default": {
      "app_id": "YOUR_APP_ID"
    },
    "work": {
      "app_id": "WORK_APP_ID"
    }
  }
}
```

### Config Loading

```go
func Load() (*Config, error)
```

1. Determines config path: `$HOME/.meta-cli/config.json`
2. Creates directory if it doesn't exist
3. If file doesn't exist, creates it with defaults
4. If file exists, reads and unmarshals JSON
5. Returns populated `*Config`

### Config Saving

```go
func (c *Config) Save() error
```

Writes the current config struct to the JSON file. Used by `pages set-default`, `auth login`, and `config set`.

---

## OS Keyring (Credentials)

**File:** `internal/auth/keyring.go`

### Platform Backends

| Platform | Backend |
|----------|---------|
| macOS | Keychain |
| Windows | Windows Credential Manager |
| Linux | Secret Service (D-Bus) / GNOME Keyring / KWallet |

### Keyring Entries

The keyring service name is `"meta-cli"`. Two entries are stored per account:

| Key Pattern | Content | Format |
|-------------|---------|--------|
| `tokens:{account}` | User token + all page tokens | JSON |
| `secret:{account}` | App Secret | Plain string |

### Store Interface

```go
type Store interface {
    GetTokens(account string) (*Tokens, error)
    SetTokens(account string, tokens *Tokens) error
    GetSecret(account string) (string, error)
    SetSecret(account string, secret string) error
    Delete(account string) error
}
```

### Token Data Format

```json
{
  "user_token": "EAAxxxxxxx...",
  "expires_at": "2026-05-06T12:00:00Z",
  "pages": {
    "123456789": {
      "name": "My Page",
      "token": "EAAxxxxxxx..."
    },
    "987654321": {
      "name": "Another Page",
      "token": "EAAxxxxxxx..."
    }
  }
}
```

### Token Access Methods

```go
func (t *Tokens) PageAccessToken(pageID string) (string, bool)
```

Returns the access token for a specific page, or `false` if not found.

---

## SQLite Database (Messages)

**File:** `internal/messenger/store.go`

### Database Location

Default: `~/.meta-cli/messages.db`

Can be overridden via the `db_path` field in config.

### Connection Settings

- Driver: `github.com/mattn/go-sqlite3`
- Journal mode: WAL (Write-Ahead Logging) for concurrent read/write performance
- Connection string: `path?_journal_mode=WAL`

### Schema

```sql
CREATE TABLE IF NOT EXISTS messages (
    id           TEXT PRIMARY KEY,     -- Facebook message ID (mid)
    psid         TEXT NOT NULL,        -- Page-scoped user ID
    page_id      TEXT NOT NULL,        -- Facebook Page ID
    text         TEXT NOT NULL,        -- Message text content
    direction    TEXT NOT NULL DEFAULT 'in',  -- 'in' or 'out'
    auto_replied INTEGER NOT NULL DEFAULT 0,  -- Was auto-replied (0/1)
    received_at  INTEGER NOT NULL      -- Unix timestamp
);

CREATE INDEX IF NOT EXISTS idx_messages_psid ON messages(psid);
CREATE INDEX IF NOT EXISTS idx_messages_page ON messages(page_id);
CREATE INDEX IF NOT EXISTS idx_messages_ts   ON messages(received_at DESC);
```

### Indexes

| Index | Column | Purpose |
|-------|--------|---------|
| `idx_messages_psid` | `psid` | Filter by user |
| `idx_messages_page` | `page_id` | Filter by page |
| `idx_messages_ts` | `received_at DESC` | Sort by most recent |

### Store Operations

```go
// Open database and initialize schema
func OpenStore(path string) (*Store, error)

// Save a message (INSERT OR IGNORE prevents duplicates)
func (s *Store) SaveMessage(m *Message) error

// List messages for a page, ordered by most recent
func (s *Store) ListMessages(pageID string, limit int) ([]Message, error)

// List recent messages for a PSID, ordered chronologically (oldest first)
func (s *Store) RecentMessages(pageID, psid string, limit int) ([]Message, error)

// Check if a message already exists (for deduplication)
func (s *Store) MessageExists(id string) bool

// Mark a message as auto-replied
func (s *Store) MarkAutoReplied(msgID string) error

// Close the database connection
func (s *Store) Close() error
```

### Message Data Structure

```go
type Message struct {
    ID          string    `json:"id"`
    PSID        string    `json:"psid"`
    PageID      string    `json:"page_id"`
    Text        string    `json:"text"`
    Direction   string    `json:"direction"`    // "in" or "out"
    AutoReplied bool      `json:"auto_replied"`
    ReceivedAt  time.Time `json:"received_at"`
}
```

### Message Sources

Messages are stored from two sources:

1. **Webhook server** - Incoming messages and message echoes received via webhook
2. **`messenger send` command** - Outgoing messages sent via the CLI

Both use `INSERT OR IGNORE` to prevent duplicate entries based on the message ID primary key.

---

## File System Layout

```
~/.meta-cli/
├── config.json       # Application configuration
├── messages.db       # SQLite message database
├── messages.db-wal   # SQLite WAL file (auto-managed)
├── messages.db-shm   # SQLite shared memory (auto-managed)
├── webhook.pid       # Daemon PID file (when running)
└── webhook.log       # Daemon log file (when running)
```

All files are stored under `~/.meta-cli/` with the directory created as `0700` (owner-only access).
