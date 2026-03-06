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
3. Select **Other** as the use case
4. Choose **None** as the app type (not "Business" — Business apps use "Facebook Login for Business" which doesn't support traditional `pages_*` scopes)
5. Give it a name (e.g. "My Page CLI") and click **Create App**

> **Important:** The app type cannot be changed after creation. If you get "Invalid Scopes" errors, create a new app with type "None".

### 2. Add Use Cases

In your app dashboard, go to **Use Cases** > **Add Use Cases** and add both:

1. **"Manage everything on your Page"** — enables `pages_manage_posts`, `pages_read_engagement`, `pages_read_user_content`, `pages_manage_engagement`, `pages_show_list`
2. **"Engage with customers on Messenger from Meta"** — enables `pages_messaging` for sending/receiving Messenger messages

After adding each use case, click **Customize** and make sure all the permissions listed are toggled on.

### 3. Add Facebook Login

1. In the app dashboard, go to **Products** > **Add Product**
2. Find **Facebook Login** (not "Facebook Login for Business") and click **Set Up**
3. Go to **Facebook Login** > **Settings**
4. Under **Valid OAuth Redirect URIs**, add `https://localhost/`
5. Click **Save Changes**

### 4. Get App ID and App Secret

1. Go to **App Settings** > **Basic**
2. Copy the **App ID**
3. Click **Show** next to **App Secret** and copy it

### 5. Connect a Facebook Page

You must be an admin of the Facebook Page you want to manage. Your pages will be automatically discovered during `auth login`.

> **Note:** In development mode, only you (the app admin) can use the app on your own pages — no app review required.

## Quick Start

```bash
# --- Auth ---
meta-cli auth login --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET
meta-cli auth status
meta-cli auth refresh

# --- Pages ---
meta-cli pages list
meta-cli pages set-default YOUR_PAGE_ID

# --- Posts ---
meta-cli post create --message "Hello from meta-cli!"
meta-cli post create --photo /path/to/image.jpg --message "Check this out!"
meta-cli post create --photo img1.jpg --photo img2.jpg --message "Album post"
meta-cli post list
meta-cli post delete POST_ID

# --- Comments ---
meta-cli comment list POST_ID
meta-cli comment reply COMMENT_ID --message "Thanks!"
meta-cli comment hide COMMENT_ID
meta-cli comment unhide COMMENT_ID
meta-cli comment delete COMMENT_ID

# --- Messenger ---
meta-cli messenger send --psid USER_PSID --message "Hello!"
meta-cli messenger list

# --- Webhook ---
meta-cli webhook serve --verify-token YOUR_TOKEN --rag-dir ./docs
meta-cli webhook subscribe
meta-cli webhook status
meta-cli webhook stop

# --- RAG ---
meta-cli rag index ./docs
meta-cli rag search "how to reset password"
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login` | Login with Facebook OAuth |
| `auth status` | Show current auth status |
| `auth refresh` | Refresh page tokens |
| `pages list` | List pages you manage |
| `pages set-default` | Set a default page for commands |
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
| `webhook subscribe` | Subscribe page to webhook fields |
| `webhook status` | Check if webhook server is running |
| `webhook stop` | Stop the webhook server |
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

Config file: `~/.meta-cli/config.json`

```json
{
  "default_account": "default",
  "default_page": "123456789",
  "graph_api_version": "v25.0",
  "webhook_port": 8080,
  "rag_dir": "./docs",
  "db_path": ""
}
```

## Permissions

The following Facebook permissions are requested during login:

- `pages_show_list` - List pages you manage
- `pages_read_engagement` - Read page engagement data
- `pages_read_user_content` - Read posts and comments
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
