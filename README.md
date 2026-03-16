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
4. Under **Valid OAuth Redirect URIs**, add `https://localhost/` (or your custom `redirect_uri` if configured)
5. Click **Save Changes**

> **Tip:** If Facebook rejects `localhost` in the App Domains field, set a custom redirect URI with a TLD:
> ```bash
> meta-cli config set redirect_uri "https://myapp.local/"
> ```
> Then add `myapp.local` to your Facebook App Domains and `https://myapp.local/` to Valid OAuth Redirect URIs.

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
meta-cli post create --video /path/to/video.mp4 --message "Watch this!"
meta-cli post create --video clip.mp4 --title "My Video" --message "Description"
meta-cli post create --video clip.mp4 --title "My Video" --thumbnail thumb.jpg --message "Description"
meta-cli post list
meta-cli post update POST_ID --message "Updated text"
meta-cli post delete POST_ID
meta-cli post create --message "Shop now!" --cta '{"type":"SHOP_NOW","value":{"link":"https://example.com/shop"}}'
meta-cli post create --message "Coming soon!" --schedule "2026-03-20 14:00"
meta-cli post create --message "Hello!" --schedule "2026-03-20 14:00" --tz "Asia/Yangon"
meta-cli post create --video clip.mp4 --message "Coming soon!" --schedule "2026-03-20 14:00"
meta-cli post create --video clip.mp4 --message "Hello!" --schedule "2026-03-20 14:00" --tz "Asia/Yangon"
meta-cli post list-scheduled

# --- Reels ---
meta-cli reel create --video /path/to/clip.mp4 --message "Check this out!"
meta-cli reel create --video clip.mp4 --title "My Reel" --message "Description"
meta-cli reel create --video clip.mp4 --message "Coming soon!" --schedule "2026-03-20 14:00"
meta-cli reel create --video clip.mp4 --message "Hello!" --schedule "2026-03-20 14:00" --tz "Asia/Yangon"

# --- Comments ---
meta-cli comment list POST_ID
meta-cli comment reply COMMENT_ID --message "Thanks!"
meta-cli comment update COMMENT_ID --message "Updated reply"
meta-cli comment hide COMMENT_ID
meta-cli comment unhide COMMENT_ID
meta-cli comment delete COMMENT_ID

# --- Insights ---
meta-cli insight page
meta-cli insight page --metric page_fans --period lifetime
meta-cli insight page --metric page_impressions,page_engaged_users --period week
meta-cli insight post POST_ID
meta-cli insight post POST_ID --metric post_impressions,post_clicks

# --- Labels ---
meta-cli label list
meta-cli label create --name "VIP"
meta-cli label delete LABEL_ID
meta-cli label assign LABEL_ID --psid USER_PSID
meta-cli label remove LABEL_ID --psid USER_PSID
meta-cli label list-by-user USER_PSID

# --- Posts (Visitor & Tagged) ---
meta-cli post list-visitor
meta-cli post list-tagged

# --- Events ---
meta-cli event list

# --- Ratings ---
meta-cli rating list
meta-cli rating summary

# --- Reactions ---
meta-cli reaction list POST_ID
meta-cli reaction list COMMENT_ID

# --- Blocked Users ---
meta-cli blocked list
meta-cli blocked add USER_ID
meta-cli blocked remove USER_ID

# --- Roles ---
meta-cli role list
meta-cli role assign USER_ID --tasks MANAGE,CREATE_CONTENT
meta-cli role remove USER_ID

# --- Messenger ---
meta-cli messenger send --psid USER_PSID --message "Hello!"
meta-cli messenger send --psid USER_PSID --image /path/to/image.jpg
meta-cli messenger send --psid USER_PSID --image https://example.com/image.jpg
meta-cli messenger send --psid USER_PSID --video /path/to/clip.mp4
meta-cli messenger send --psid USER_PSID --message "Hello!" --tag HUMAN_AGENT
meta-cli messenger send --psid USER_PSID --message "Pick one" --quick-reply "Yes" --quick-reply "No"
meta-cli messenger send-template --psid USER_PSID --json '{"template_type":"button","text":"Hello","buttons":[]}'
meta-cli messenger list
meta-cli messenger history --psid USER_PSID
meta-cli messenger conversations

# --- Messenger Profile ---
meta-cli messenger profile get
meta-cli messenger profile set-greeting "Welcome!"
meta-cli messenger profile set-get-started GET_STARTED
meta-cli messenger profile set-menu --file menu.json
meta-cli messenger profile set-ice-breakers --file icebreakers.json
meta-cli messenger profile delete --field greeting

# --- Leads ---
meta-cli lead create-form --json '{"name":"Contact Form"}'
meta-cli lead list FORM_ID

# --- Config ---
meta-cli config set verify_token YOUR_TOKEN
meta-cli config get verify_token
meta-cli config list

# --- Webhook ---
meta-cli webhook serve --verify-token YOUR_TOKEN
meta-cli webhook subscribe
meta-cli webhook status
meta-cli webhook stop

# --- Auto-Reply (with OpenClaw) ---
meta-cli config set auto_reply true
meta-cli config set hooks_endpoint http://127.0.0.1:18789/hooks/agent
meta-cli config set hooks_token YOUR_TOKEN
meta-cli config set rag_dir ./knowledge-base
meta-cli webhook serve --auto-reply --daemon

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
| `post create` | Create a text, photo, video, or link post |
| `post update` | Update a post's message |
| `post edit` | Edit a post's message |
| `post delete` | Delete a post |
| `post list-scheduled` | List scheduled (unpublished) posts |
| `reel create` | Publish a reel (short-form video) |
| `comment list` | List comments on a post |
| `comment reply` | Reply to a comment |
| `comment update` | Update a comment's message |
| `comment hide` | Hide a comment |
| `comment unhide` | Unhide a comment |
| `comment delete` | Delete a comment |
| `insight page` | Show page-level insights |
| `insight post` | Show post-level insights |
| `label list` | List all custom labels for the page |
| `label create` | Create a new custom label |
| `label delete` | Delete a custom label |
| `label assign` | Assign a label to a user |
| `label remove` | Remove a label from a user |
| `label list-by-user` | List labels assigned to a user |
| `pages info` | Display page information |
| `post list-visitor` | List visitor posts on the page |
| `post list-tagged` | List posts where the page is tagged |
| `post create --backdate` | Create backdated posts |
| `post create --targeting` | Create audience-targeted posts |
| `post create --place` | Create location-tagged posts |
| `post create --cta` | Add call-to-action button to posts |
| `event list` | List page events |
| `rating list` | List page ratings and reviews |
| `rating summary` | Show overall page rating |
| `reaction list` | List reactions on a post or comment |
| `blocked list` | List blocked users |
| `blocked add` | Block a user |
| `blocked remove` | Unblock a user |
| `role list` | List users with page access |
| `role assign` | Assign roles to a user |
| `role remove` | Remove a user's page access |
| `comment private-reply` | Send a private Messenger reply to a comment |
| `messenger send` | Send a Messenger message (text, attachment, tagged, quick replies) |
| `messenger send-template` | Send a template message |
| `messenger list` | List stored messages |
| `messenger history` | List conversation history with a user |
| `messenger conversations` | List Messenger conversations from API |
| `messenger profile get` | Get Messenger profile settings |
| `messenger profile set-greeting` | Set the Messenger greeting text |
| `messenger profile set-get-started` | Set the Get Started button payload |
| `messenger profile set-menu` | Set the persistent menu |
| `messenger profile set-ice-breakers` | Set ice breaker conversation starters |
| `messenger profile delete` | Delete a Messenger profile field |
| `lead create-form` | Create a lead generation form |
| `lead list` | List leads from a form |
| `config set` | Set a config value |
| `config get` | Get a config value |
| `config list` | List all config values |
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
  "db_path": "",
  "verify_token": "",
  "redirect_uri": "",
  "debounce_seconds": 3,
  "hooks_endpoint": "",
  "hooks_token": "",
  "auto_reply": false,
  "prompt_template": ""
}
```

| Field | Description |
|-------|-------------|
| `redirect_uri` | Custom OAuth redirect URI (default: `https://localhost/`) |
| `debounce_seconds` | Debounce window for message batching (default: 3) |
| `hooks_endpoint` | OpenClaw webhook endpoint URL |
| `hooks_token` | OpenClaw webhook auth token |
| `auto_reply` | Enable auto-reply pipeline (true/false) |
| `prompt_template` | Go template for agent prompt |

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
  pages/                   Page listing + info
  posts/                   Post CRUD + visitor/tagged posts
  comments/                Comment management + private reply
  insights/                Page and post analytics
  labels/                  Custom label management (CRUD + user assignment)
  messenger/               Send messages + attachments + templates + profile + SQLite store + webhook
  events/                  Page events listing
  ratings/                 Page ratings and reviews
  reactions/               Reaction listing on posts/comments
  blocked/                 Blocked user management
  roles/                   User/role management
  leads/                   Lead generation forms and leads
  rag/                     TF-IDF search over markdown documents
  debounce/                Message debouncing (per-user timer)
  hooks/                   OpenClaw /hooks/agent caller
```

## License

MIT
