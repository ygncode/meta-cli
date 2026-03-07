# Architecture

## High-Level Overview

meta-cli follows the standard Go project layout with a clear separation between CLI command definitions and internal business logic.

```
meta-cli/
├── cmd/meta/main.go          # Binary entry point
├── cmd_impl/                  # Cobra command definitions
│   ├── root.go                # Root command, global flags, context setup
│   ├── auth.go                # auth login/status/refresh
│   ├── pages.go               # pages list/set-default
│   ├── posts.go               # post list/create/delete
│   ├── comments.go            # comment list/reply/hide/unhide/delete
│   ├── messenger.go           # messenger send/list
│   ├── webhook.go             # webhook serve/subscribe/status/stop
│   ├── config.go              # config set/get/list
│   └── rag.go                 # rag index/search
├── internal/                  # Internal packages (not importable externally)
│   ├── auth/                  # OAuth flow + OS keyring storage
│   ├── config/                # JSON config management
│   ├── graph/                 # Meta Graph API HTTP client
│   ├── output/                # Table/JSON/TSV output formatting
│   ├── pages/                 # Page listing service
│   ├── posts/                 # Post CRUD operations
│   ├── comments/              # Comment management service
│   ├── messenger/             # Messaging + SQLite store + webhook handler
│   ├── rag/                   # TF-IDF document search
│   └── daemon/                # Background process management
├── go.mod                     # Module definition
├── go.sum                     # Dependency checksums
└── Makefile                   # Build automation
```

## Execution Flow

### Startup Sequence

```
main()
  └── cmd_impl.Execute()
        └── rootCmd.Execute()              # Cobra parses args
              └── PersistentPreRunE()       # Runs before ANY command
                    ├── config.Load()       # Load ~/.meta-cli/config.json
                    ├── Resolve output format (--json / --plain / table)
                    ├── Resolve account (--account flag or config default)
                    ├── Resolve page ID (--page flag or config default)
                    ├── Create KeyringStore  # OS keyring access
                    ├── Create Printer       # Output formatter
                    └── Inject RootCtx       # Store in cobra.Command context
```

### Request Pipeline

When a command that requires API access runs:

```
Command Handler (e.g., post list)
  └── requirePageClient(cmd)
        ├── requireTokens(cmd)
        │     ├── GetCtx(cmd)              # Retrieve RootCtx
        │     └── Store.GetTokens(account) # Load from OS keyring
        ├── Resolve page token from Tokens.Pages[pageID]
        └── Create graph.Client with page token
              └── graph.Client.Get/Post/Delete()
                    └── HTTP request to https://graph.facebook.com/v25.0/...
                          └── Parse response / handle errors
```

## Package Dependency Graph

```
cmd/meta/main.go
  └── cmd_impl
        ├── internal/config      (configuration loading)
        ├── internal/auth        (token management)
        ├── internal/graph       (API client)
        ├── internal/output      (formatting)
        ├── internal/pages       (page service)
        ├── internal/posts       (post service)
        ├── internal/comments    (comment service)
        ├── internal/messenger   (messaging + webhooks + SQLite)
        ├── internal/rag         (document search)
        └── internal/daemon      (process management)
```

Dependencies flow strictly downward - `cmd_impl` depends on `internal/*` packages, but `internal/*` packages do not depend on each other (with the exception of `messenger` using `graph.Client`).

## Context System

The CLI uses Go's `context.Context` to pass shared state through the command tree. The `RootCtx` struct is the central context object:

```go
type RootCtx struct {
    Config  *config.Config    // Loaded configuration
    Store   auth.Store        // OS keyring interface
    Tokens  *auth.Tokens      // User + page tokens (lazy-loaded)
    Client  *graph.Client     // Graph API client (lazy-loaded)
    Printer *output.Printer   // Output formatter
    Account string            // Active account name
    PageID  string            // Active page ID
}
```

`RootCtx` is injected during `PersistentPreRunE` and retrieved by commands via `GetCtx(cmd)`. The `Tokens` and `Client` fields are lazily populated by `requireTokens()` and `requirePageClient()` helper functions - only commands that need API access trigger token loading.

## Command Registration Pattern

All commands use Go's `init()` functions for self-registration:

```go
// In each command file (e.g., posts.go)
func init() {
    parentCmd := &cobra.Command{
        Use:   "post",
        Short: "Manage posts",
        Aliases: []string{"posts"},
    }
    parentCmd.AddCommand(postListCmd())
    parentCmd.AddCommand(postCreateCmd())
    parentCmd.AddCommand(postDeleteCmd())
    rootCmd.AddCommand(parentCmd)
}
```

When the `cmd_impl` package is imported by `main.go`, all `init()` functions run automatically, building the complete command tree before `Execute()` is called.

## Service Layer

Each domain area has a service struct that accepts a `graph.Client` and provides high-level operations:

```
graph.Client (HTTP)
  ├── posts.Service      → List, CreateText, CreatePhoto, CreatePhotos, CreateLink, Delete
  ├── comments.Service   → List, Reply, SetHidden, Delete
  ├── pages.Service      → List
  └── messenger.Service  → Send, SendTyping, SubscribeWebhook
```

Services encapsulate the specific Graph API endpoints, request/response formats, and business logic for each domain.

## Output System

The `output.Printer` supports three formats, selected by global flags:

| Format | Flag | Implementation |
|--------|------|----------------|
| Table | (default) | `olekukonko/tablewriter` - ASCII tables |
| JSON | `--json` | `encoding/json` - JSON arrays/objects |
| TSV | `--plain` | Tab-separated values for shell scripting |

The printer uses Go's `reflect` package to extract fields from structs based on their `json` struct tags, providing automatic column discovery without manual mapping.

## Error Handling Strategy

- **Graph API errors** are wrapped in `graph.APIError` with HTTP status code and Graph-specific error details (code, message, type, fbtrace_id)
- Helper functions `IsTokenExpired()` (code 190) and `IsPermissionDenied()` (code 200/10) enable specific error handling
- Command-level errors bubble up through Cobra's `RunE` handlers and are printed to stderr
- The process exits with code 1 on any unhandled error

## External Dependencies

| Dependency | Version | Purpose |
|-----------|---------|---------|
| `github.com/spf13/cobra` | v1.10.2 | CLI framework (commands, flags, help) |
| `github.com/zalando/go-keyring` | v0.2.6 | OS keyring for secure token storage |
| `github.com/mattn/go-sqlite3` | v1.14.34 | SQLite driver for message storage |
| `github.com/olekukonko/tablewriter` | v0.0.5 | ASCII table rendering |

All other functionality uses the Go standard library (`net/http`, `encoding/json`, `crypto/hmac`, `database/sql`, etc.).
