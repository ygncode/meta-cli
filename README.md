# meta-cli

CLI for managing Facebook Pages and Messenger via Meta Graph API.

## Install

```bash
go install github.com/ygncode/meta-cli/cmd/meta@latest
```

Or build from source:

```bash
git clone https://github.com/ygncode/meta-cli.git
cd meta-cli
make build
```

## Prerequisites

You need a Facebook App to get your **App ID** and **App Secret**. In development mode (default), no app review is needed — you can manage your own pages immediately.

### 1. Create a Facebook App

1. Go to [Meta for Developers](https://developers.facebook.com/) and log in
2. Click **My Apps** > **Create App**
3. Select **Other** as the use case, then choose **Business** as the app type
4. Give it a name (e.g. "My Page CLI") and click **Create App**

### 2. Get App ID and App Secret

1. In your app dashboard, go to **App Settings** > **Basic**
2. Copy the **App ID**
3. Click **Show** next to **App Secret** and copy it

### 3. Add Facebook Login

1. In the app dashboard, click **Add Product**
2. Find **Facebook Login** and click **Set Up**
3. Go to **Facebook Login** > **Settings**
4. Under **Valid OAuth Redirect URIs**, add `https://localhost/`
5. Click **Save Changes**

### 4. Connect a Facebook Page

1. Go to **App Settings** > **Advanced**
2. Under **App Page**, connect the Facebook Page you want to manage
3. Alternatively, the page will be automatically discovered during `auth login` if you are an admin of the page

> **Note:** In development mode, only your own Facebook account can use the app. This is fine for personal use — no app review required.

## Quick Start

```bash
# Login with your Facebook App credentials
meta-cli auth login --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET

# List your pages
meta-cli pages list

# Set default page in config
# Edit ~/.config/meta-cli/config.json and set "default_page": "YOUR_PAGE_ID"

# Create a text post
meta-cli post create --message "Hello from meta-cli!"

# List recent posts
meta-cli post list --page YOUR_PAGE_ID

# List comments on a post
meta-cli comment list POST_ID

# Send a Messenger message
meta-cli messenger send --psid USER_PSID --message "Hello!"

# Start webhook server with RAG auto-reply
meta-cli webhook serve --verify-token YOUR_TOKEN --rag-dir ./docs

# Search RAG documents
meta-cli rag search "how to reset password"
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login` | Login with Facebook OAuth |
| `auth status` | Show current auth status |
| `auth refresh` | Refresh page tokens |
| `pages list` | List pages you manage |
| `post list` | List recent posts |
| `post create` | Create a text, photo, or link post |
| `post delete` | Delete a post |
| `comment list` | List comments on a post |
| `comment reply` | Reply to a comment |
| `comment hide` | Hide a comment |
| `comment unhide` | Unhide a comment |
| `comment delete` | Delete a comment |
| `messenger send` | Send a Messenger message |
| `messenger list` | List stored messages |
| `webhook serve` | Start webhook HTTP server |
| `rag index` | Show index stats for documents |
| `rag search` | Search documents by query |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |
| `--plain` | Output as TSV |
| `--page` | Page ID to operate on |
| `--account` | Account name (default: "default") |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `META_VERIFY_TOKEN` | Webhook verification token |

## Config

Config file: `~/.config/meta-cli/config.json`

```json
{
  "default_account": "default",
  "default_page": "123456789",
  "graph_api_version": "v21.0",
  "webhook_port": 8080,
  "rag_dir": "./docs",
  "db_path": ""
}
```

## Permissions

The following Facebook permissions are requested during login:

- `pages_show_list` - List pages you manage
- `pages_read_engagement` - Read page engagement data
- `pages_manage_posts` - Create and delete posts
- `pages_manage_engagement` - Manage comments
- `pages_messaging` - Send and receive Messenger messages
- `pages_manage_metadata` - Manage page metadata
- `public_profile` - Access public profile

## Architecture

```
cmd/meta/main.go          Entry point
cmd_impl/                  Cobra command definitions
internal/
  auth/                    OAuth flow + OS keyring storage
  config/                  JSON config management
  graph/                   Meta Graph API HTTP client
  output/                  Table/JSON/TSV output formatting
  pages/                   Page listing
  posts/                   Post CRUD operations
  comments/                Comment management
  messenger/               Send messages + SQLite store + webhook handler
  rag/                     TF-IDF search over markdown documents
```

## License

MIT
