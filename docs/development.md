# Development Guide

## Prerequisites

- Go 1.25.5 or later
- CGO enabled (required for SQLite via `go-sqlite3`)
- A Facebook App with appropriate permissions (see [README](../README.md))

## Building

### From Source

```bash
git clone https://github.com/ygncode/meta-cli.git
cd meta-cli
make build
```

This produces the `meta-cli` binary in the project root.

### Install to GOPATH

```bash
make install
```

### Available Make Targets

| Target | Command | Description |
|--------|---------|-------------|
| `build` | `go build -o meta-cli ./cmd/meta` | Compile binary |
| `install` | `go install ./cmd/meta` | Install to `$GOPATH/bin` |
| `test` | `go test ./...` | Run all tests |
| `lint` | `go vet ./...` | Run static analysis |
| `tidy` | `go mod tidy` | Clean up dependencies |
| `clean` | `rm -f meta-cli` | Remove compiled binary |

## Project Structure

```
meta-cli/
├── cmd/
│   └── meta/
│       └── main.go              # Binary entry point (minimal)
├── cmd_impl/                     # Cobra command implementations
│   ├── root.go                   # Root command + context setup
│   ├── auth.go                   # auth login/status/refresh
│   ├── pages.go                  # pages list/set-default
│   ├── posts.go                  # post list/create/delete
│   ├── comments.go               # comment list/reply/hide/unhide/delete
│   ├── messenger.go              # messenger send/list
│   ├── webhook.go                # webhook serve/subscribe/status/stop
│   └── rag.go                    # rag index/search
├── internal/                     # Internal packages
│   ├── auth/                     # OAuth flow + keyring storage
│   │   ├── auth.go               # OAuth URL generation, code exchange, token extension
│   │   ├── tokens.go             # Token types and serialization
│   │   └── keyring.go            # OS keyring Store implementation
│   ├── config/                   # Configuration
│   │   ├── config.go             # Load/save logic
│   │   └── types.go              # Config + Account structs
│   ├── graph/                    # Meta Graph API client
│   │   ├── client.go             # HTTP client (Get/Post/PostMultipart/Delete)
│   │   └── errors.go             # GraphError, APIError types
│   ├── output/                   # Output formatting
│   │   └── printer.go            # Table/JSON/TSV printer
│   ├── pages/                    # Page listing
│   │   └── service.go            # Pages service
│   ├── posts/                    # Post management
│   │   └── service.go            # Post CRUD service
│   ├── comments/                 # Comment management
│   │   └── service.go            # Comment service
│   ├── messenger/                # Messaging system
│   │   ├── service.go            # Send messages, subscribe webhooks
│   │   ├── store.go              # SQLite message store
│   │   ├── webhook.go            # Webhook HTTP handler
│   │   └── types.go              # Message, webhook payload types
│   ├── rag/                      # Document search
│   │   └── (TF-IDF implementation)
│   └── daemon/                   # Process management
│       └── daemon.go             # PID files, process signals
├── go.mod                        # Module definition
├── go.sum                        # Dependency checksums
├── Makefile                      # Build automation
└── README.md                     # User-facing documentation
```

## Adding a New Command

### 1. Create the Service (if needed)

If the command interacts with the Graph API, create a service in `internal/`:

```go
// internal/myfeature/service.go
package myfeature

import (
    "context"
    "github.com/ygncode/meta-cli/internal/graph"
)

type Service struct {
    client *graph.Client
}

func NewService(client *graph.Client) *Service {
    return &Service{client: client}
}

func (s *Service) DoSomething(ctx context.Context) error {
    // Use s.client.Get/Post/Delete
    return nil
}
```

### 2. Create the Command

Add a new file in `cmd_impl/` or add to an existing one:

```go
// cmd_impl/myfeature.go
package cmd_impl

import "github.com/spf13/cobra"

func init() {
    parentCmd := &cobra.Command{
        Use:   "myfeature",
        Short: "Description of my feature",
    }
    parentCmd.AddCommand(mySubCmd())
    rootCmd.AddCommand(parentCmd)
}

func mySubCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "action",
        Short: "Do something",
        RunE: func(cmd *cobra.Command, args []string) error {
            rctx, err := requirePageClient(cmd) // or requireTokens(cmd) if no page needed
            if err != nil {
                return err
            }

            svc := myfeature.NewService(rctx.Client)
            result, err := svc.DoSomething(cmd.Context())
            if err != nil {
                return err
            }

            rctx.Printer.PrintOne(result)
            return nil
        },
    }
    return cmd
}
```

### 3. Key Patterns to Follow

- Use `requireTokens(cmd)` for commands that need authentication but not a specific page
- Use `requirePageClient(cmd)` for commands that need authentication + page token
- Use `rctx.Printer` for output (respects `--json` and `--plain` flags automatically)
- Return errors from `RunE` - Cobra handles printing them
- Use `cmd.Context()` for the context passed to service methods

## Testing

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./internal/graph/...

# Run with verbose output
go test -v ./...
```

The `graph.NewWithHTTPClient()` constructor and `auth.MemStore` are designed for testing - they allow injecting mock HTTP clients and in-memory credential stores.

## Dependencies

| Package | Purpose | Why This One |
|---------|---------|-------------|
| `spf13/cobra` | CLI framework | Industry standard for Go CLIs |
| `zalando/go-keyring` | OS keyring access | Cross-platform, well-maintained |
| `mattn/go-sqlite3` | SQLite driver | Most mature Go SQLite driver |
| `olekukonko/tablewriter` | Table formatting | Simple API, good output |

### CGO Note

`go-sqlite3` requires CGO. On macOS, this works out of the box with Xcode Command Line Tools. On Linux, ensure `gcc` is installed:

```bash
# Debian/Ubuntu
sudo apt-get install gcc

# Alpine
apk add gcc musl-dev
```

## Configuration During Development

The CLI reads config from `~/.meta-cli/config.json`. You can create this file manually or let `auth login` create it:

```bash
# Manually create config directory
mkdir -p ~/.meta-cli

# Create minimal config
echo '{"graph_api_version":"v25.0"}' > ~/.meta-cli/config.json
```

## Debugging

### View API requests

The Graph API client uses Go's standard `net/http` client. To inspect requests, you can set the `HTTPS_PROXY` environment variable:

```bash
HTTPS_PROXY=http://localhost:8888 meta-cli post list
```

### Check stored tokens

```bash
meta-cli auth status --json
```

### View webhook logs (daemon mode)

```bash
tail -f ~/.meta-cli/webhook.log
```

### Check database contents

```bash
sqlite3 ~/.meta-cli/messages.db "SELECT * FROM messages ORDER BY received_at DESC LIMIT 10;"
```
