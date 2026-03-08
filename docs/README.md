# meta-cli Documentation

Detailed documentation for **meta-cli** - a Go CLI for managing Facebook Pages and Messenger via the Meta Graph API.

## Table of Contents

| Document | Description |
|----------|-------------|
| [Architecture](./architecture.md) | System architecture, package layout, and design decisions |
| [Authentication Flow](./authentication.md) | OAuth 2.0 flow, token lifecycle, and credential storage |
| [Commands Reference](./commands.md) | Complete reference for all CLI commands, flags, and options |
| [Webhook System](./webhooks.md) | Webhook server, daemon mode, message processing pipeline |
| [Data Storage](./storage.md) | Configuration, SQLite database, and OS keyring details |
| [Graph API Client](./graph-api.md) | Meta Graph API integration, endpoints, and error handling |
| [Auto-Reply Guide](./auto-reply.md) | OpenClaw integration for automatic Messenger replies |
| [Development Guide](./development.md) | Building, testing, and contributing to the project |

## Quick Overview

meta-cli is structured as a standard Go CLI application using the [Cobra](https://github.com/spf13/cobra) framework. It communicates with the [Meta Graph API](https://developers.facebook.com/docs/graph-api/) to manage Facebook Pages, posts, comments, labels, and Messenger conversations.

```
User --> meta-cli (Cobra) --> Meta Graph API
                |                (Posts, Comments, Labels,
                |                 Messenger, Webhooks)
                |
                +--> OS Keyring (credentials)
                +--> SQLite DB (message history)
                +--> JSON config (~/.meta-cli/config.json)
```

### Key Design Principles

- **Secure credential storage** - Tokens are stored in the OS keyring, never in plain files
- **Multi-account support** - Manage multiple Facebook accounts from a single installation
- **Flexible output** - Table, JSON, or TSV output for scripting integration
- **Daemon-capable webhooks** - Run webhook server in foreground or as a background daemon
- **Minimal dependencies** - Only 4 direct dependencies beyond the Go standard library
