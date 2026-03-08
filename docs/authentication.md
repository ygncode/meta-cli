# Authentication Flow

## Overview

meta-cli uses Facebook's OAuth 2.0 flow to obtain access tokens. Tokens are stored securely in the OS keyring (macOS Keychain, Windows Credential Manager, or Linux Secret Service/D-Bus).

## OAuth 2.0 Flow

### Step-by-step Process

```
┌─────────────────────────────────────────────────────────────┐
│ 1. meta-cli auth login --app-id XXX --app-secret YYY       │
│    └── Generates OAuth URL and prints it                    │
│                                                             │
│ 2. User opens URL in browser                                │
│    └── https://www.facebook.com/v25.0/dialog/oauth?...      │
│                                                             │
│ 3. User authenticates and grants permissions                │
│    └── Facebook redirects to redirect_uri/?code=ABC          │
│        (default: https://localhost/)                         │
│                                                             │
│ 4. User pastes the redirect URL back into the CLI           │
│    └── CLI extracts the authorization code from URL          │
│                                                             │
│ 5. Exchange code for short-lived token                      │
│    └── POST /oauth/access_token?code=ABC&client_id=...      │
│                                                             │
│ 6. Extend to long-lived token                               │
│    └── POST /oauth/access_token?grant_type=fb_exchange_token│
│                                                             │
│ 7. Fetch page access tokens                                 │
│    └── GET /me/accounts?access_token=LONG_LIVED_TOKEN       │
│                                                             │
│ 8. Store in OS keyring                                      │
│    └── User token + all page tokens saved securely          │
└─────────────────────────────────────────────────────────────┘
```

### Required Scopes

The following Facebook permissions are requested during the OAuth flow:

| Scope | Purpose |
|-------|---------|
| `pages_show_list` | List pages the user manages |
| `pages_read_engagement` | Read page engagement metrics (likes, shares) |
| `pages_read_user_content` | Read posts and comments on pages |
| `pages_manage_posts` | Create and delete posts |
| `pages_manage_engagement` | Manage comments (reply, hide, delete) |
| `pages_messaging` | Send and receive Messenger messages |
| `pages_manage_metadata` | Manage page metadata and subscriptions |
| `public_profile` | Access basic profile information |

## Token Types

### User Access Token

- Obtained by exchanging the OAuth authorization code
- Short-lived token (~1 hour) is extended to long-lived (~60 days)
- Used to list manageable pages and fetch page tokens
- Stored in OS keyring under key `tokens:{account}`

### Page Access Tokens

- Obtained via `GET /me/accounts` using the user token
- Each page has its own access token
- These tokens do not expire as long as the user token is valid
- Used for all page-specific operations (posts, comments, messages)
- Stored as a map within the `Tokens` struct in the keyring

## Token Storage

### Keyring Implementation

```go
// internal/auth/keyring.go
type KeyringStore struct{}

// Service name used for all keyring entries
const serviceName = "meta-cli"

// Keys stored per account:
// - "tokens:{account}"  → JSON-serialized Tokens struct
// - "secret:{account}"  → App Secret (for webhook signature validation)
```

The `auth.Store` interface abstracts the storage backend:

```go
type Store interface {
    GetTokens(account string) (*Tokens, error)
    SetTokens(account string, tokens *Tokens) error
    GetSecret(account string) (string, error)
    SetSecret(account string, secret string) error
    Delete(account string) error
}
```

Two implementations exist:
- `KeyringStore` - Production implementation using the OS keyring
- `MemStore` - In-memory implementation for testing

### Token Data Structure

```go
type Tokens struct {
    UserToken string               `json:"user_token"`
    ExpiresAt time.Time            `json:"expires_at"`
    Pages     map[string]PageToken `json:"pages"`
}

type PageToken struct {
    Name  string `json:"name"`
    Token string `json:"token"`
}
```

### Serialization

Tokens are serialized to JSON before storing in the keyring and deserialized on retrieval. The entire `Tokens` struct (user token + all page tokens) is stored as a single keyring entry.

## Token Refresh

```
meta-cli auth refresh
  └── Loads existing user token from keyring
       └── Calls FetchPageTokens() with user token
            └── Updates page tokens in keyring
```

This refreshes page access tokens without requiring a new OAuth flow. Useful when new pages are added to the account or when page tokens need rotation.

## Multi-Account Support

Multiple Facebook accounts are supported through the `--account` flag:

```bash
# Login with different accounts
meta-cli auth login --app-id APP1 --app-secret SECRET1 --account work
meta-cli auth login --app-id APP2 --app-secret SECRET2 --account personal

# Use specific account
meta-cli post list --account work
meta-cli post list --account personal
```

Each account gets separate keyring entries (`tokens:work`, `tokens:personal`) and can have a different App ID configured in `config.json`.

## Security Considerations

- **No plaintext storage** - Tokens are never written to disk files; they're stored in the OS-level encrypted keyring
- **App Secret protection** - The app secret is also stored in the keyring, used only for webhook HMAC signature validation
- **HTTPS only** - All communication with the Meta Graph API uses HTTPS
- **Redirect URI** - Defaults to `https://localhost/` as the redirect URI, which is a standard pattern for CLI OAuth flows. Can be customized via `config set redirect_uri` for environments where `localhost` is not accepted by Facebook App settings
- **Scope minimization** - Only requests the specific permissions needed for the CLI's functionality
