# Commands Reference

## Command Tree

```
meta-cli
├── auth
│   ├── login          Login with Facebook OAuth
│   ├── status         Show current auth status
│   └── refresh        Refresh page tokens
├── pages
│   ├── list           List pages you manage
│   └── set-default    Set a default page for commands
├── post (alias: posts)
│   ├── list           List recent posts
│   ├── create         Create a text, photo, video, or link post
│   ├── update         Update a post's message
│   ├── edit           Edit a post's message
│   ├── delete         Delete a post
│   └── list-scheduled List scheduled (unpublished) posts
├── comment (alias: comments)
│   ├── list           List comments on a post
│   ├── reply          Reply to a comment
│   ├── update         Update a comment's message
│   ├── hide           Hide a comment
│   ├── unhide         Unhide a comment
│   └── delete         Delete a comment
├── insight (alias: insights)
│   ├── page           Show page-level insights
│   └── post           Show post-level insights
├── label (alias: labels)
│   ├── list           List all custom labels for the page
│   ├── create         Create a new custom label
│   ├── delete         Delete a custom label
│   ├── assign         Assign a label to a user
│   ├── remove         Remove a label from a user
│   └── list-by-user   List labels assigned to a user
├── messenger
│   ├── send           Send a Messenger message
│   ├── list           List stored messages
│   └── history        List conversation history with a user
├── config
│   ├── set            Set a config value
│   ├── get            Get a config value
│   └── list           List all config values
├── webhook
│   ├── serve          Start webhook HTTP server
│   ├── subscribe      Subscribe page to webhook fields
│   ├── status         Check if webhook server is running
│   └── stop           Stop the webhook server
└── rag
    ├── index          Show index stats for documents
    └── search         Search documents by query
```

## Global Flags

These flags are available on all commands:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |
| `--plain` | bool | `false` | Output as TSV (tab-separated values) |
| `--page` | string | from config | Page ID to operate on |
| `--account` | string | `"default"` | Account name for multi-account setups |

## Auth Commands

### `auth login`

Authenticate with Facebook OAuth and store tokens securely.

```bash
meta-cli auth login --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--app-id` | string | Yes | Facebook App ID (can be stored in config) |
| `--app-secret` | string | Yes | Facebook App Secret |

**Process:**
1. Prints an OAuth URL for the user to open in a browser
2. Prompts the user to paste the redirect URL
3. Exchanges the authorization code for tokens
4. Extends the short-lived token to long-lived
5. Fetches page access tokens for all managed pages
6. Stores everything in the OS keyring

**Requires:** Nothing (initial setup command)

---

### `auth status`

Display the current authentication status and managed pages.

```bash
meta-cli auth status
meta-cli auth status --json
```

**Output:** Account name, list of pages with IDs and names. The default page is marked with `*`.

**Requires:** Previously completed `auth login`

---

### `auth refresh`

Refresh page access tokens using the existing user token.

```bash
meta-cli auth refresh
```

**Output:** Count of refreshed pages.

**Requires:** Previously completed `auth login`

---

## Pages Commands

### `pages list`

List all Facebook Pages the authenticated user can manage.

```bash
meta-cli pages list
meta-cli pages list --json
```

**Output:** Table of pages with ID and name.

**Requires:** Authentication

---

### `pages set-default`

Set a default page so you don't need `--page` on every command.

```bash
meta-cli pages set-default 123456789
```

| Argument | Required | Description |
|----------|----------|-------------|
| `page-id` | Yes | The page ID to set as default |

**Behavior:** Updates `default_page` in `~/.meta-cli/config.json`.

**Requires:** Nothing

---

## Post Commands

### `post list`

List recent posts on the page.

```bash
meta-cli post list
meta-cli post list --limit 20
meta-cli post list --json
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `10` | Number of posts to fetch |

**Output:** Table with post ID, message preview, creation time, permalink, and engagement metrics (likes, comments, shares).

**Requires:** Authentication + page

---

### `post create`

Create a new post on the page. Supports text, photo, multi-photo album, video, and link posts.

```bash
# Text post
meta-cli post create --message "Hello world!"

# Single photo post
meta-cli post create --photo /path/to/image.jpg --message "Check this out!"

# Multi-photo album post
meta-cli post create --photo img1.jpg --photo img2.jpg --message "Album post"

# Video post
meta-cli post create --video /path/to/video.mp4 --message "Watch this!"
meta-cli post create --video clip.mp4 --title "My Video" --message "Description"
meta-cli post create --video clip.mp4 --title "My Video" --thumbnail thumb.jpg --message "Description"

# Link post
meta-cli post create --link https://example.com --message "Interesting link"

# Scheduled post
meta-cli post create --message "Coming soon!" --schedule "2026-03-20 14:00"
meta-cli post create --message "Hello!" --schedule "2026-03-20 14:00" --tz "Asia/Yangon"
meta-cli post create --video clip.mp4 --message "Coming soon!" --schedule "2026-03-20 14:00"
meta-cli post create --video clip.mp4 --message "Hello!" --schedule "2026-03-20 14:00" --tz "Asia/Yangon"
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--message` | string | No* | Post text content |
| `--photo` | string[] | No* | Path to image file (repeatable for albums) |
| `--video` | string | No* | Path to video file |
| `--title` | string | No | Video title (requires `--video`) |
| `--thumbnail` | string | No | Path to thumbnail image (requires `--video`) |
| `--link` | string | No* | URL to share |
| `--schedule` | string | No | Schedule for future publishing (format: `"YYYY-MM-DD HH:MM"`) |
| `--tz` | string | No | Timezone for `--schedule` (e.g. `"Asia/Yangon"`), defaults to local |

*At least one of `--message`, `--photo`, `--video`, or `--link` is required.

**Post type resolution:**
1. If `--video` → Video post (multipart upload to `/{page-id}/videos`)
2. If multiple `--photo` flags → Album post (photos uploaded as unpublished, then attached)
3. If single `--photo` → Photo post (multipart upload)
4. If `--link` (no photos/video) → Link post
5. If only `--message` → Text post

**Output:** Created post details (ID, permalink).

**Requires:** Authentication + page

---

### `post update`

Update the message text of an existing post.

```bash
meta-cli post update POST_ID --message "Updated text"
meta-cli post update POST_ID -m "Updated text"
```

| Argument | Required | Description |
|----------|----------|-------------|
| `post-id` | Yes | The post ID to update |

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `-m`, `--message` | string | Yes | New message text |

**Note:** You can update the message of text, link, and photo posts. However, you cannot change the photo itself on a photo post — only the caption/message.

**Requires:** Authentication + page

---

### `post delete`

Delete a post from the page.

```bash
meta-cli post delete POST_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `post-id` | Yes | The post ID to delete |

**Requires:** Authentication + page

---

### `post list-scheduled`

List scheduled (unpublished) posts.

```bash
meta-cli post list-scheduled
meta-cli post list-scheduled --limit 20
meta-cli post list-scheduled --json
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `10` | Number of scheduled posts to fetch |

**Output:** Table with post ID, message, scheduled publish time, and creation time.

**Requires:** Authentication + page

---

## Comment Commands

### `comment list`

List comments on a specific post.

```bash
meta-cli comment list POST_ID
meta-cli comment list POST_ID --limit 50
```

| Argument | Required | Description |
|----------|----------|-------------|
| `post-id` | Yes | The post ID to list comments for |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `25` | Number of comments to fetch |

**Output:** Table with comment ID, author, message, timestamp, and like count.

**Requires:** Authentication + page

---

### `comment reply`

Reply to a comment.

```bash
meta-cli comment reply COMMENT_ID --message "Thanks for the feedback!"
meta-cli comment reply COMMENT_ID -m "Thanks!"
```

| Argument | Required | Description |
|----------|----------|-------------|
| `comment-id` | Yes | The comment ID to reply to |

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `-m`, `--message` | string | Yes | Reply text |

**Output:** Created reply ID.

**Requires:** Authentication + page

---

### `comment update`

Update the message text of an existing comment.

```bash
meta-cli comment update COMMENT_ID --message "Updated comment text"
meta-cli comment update COMMENT_ID -m "Updated comment text"
```

| Argument | Required | Description |
|----------|----------|-------------|
| `comment-id` | Yes | The comment ID to update |

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `-m`, `--message` | string | Yes | New message text |

**Requires:** Authentication + page

---

### `comment hide`

Hide a comment from public view.

```bash
meta-cli comment hide COMMENT_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `comment-id` | Yes | The comment ID to hide |

**Requires:** Authentication + page

---

### `comment unhide`

Unhide a previously hidden comment.

```bash
meta-cli comment unhide COMMENT_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `comment-id` | Yes | The comment ID to unhide |

**Requires:** Authentication + page

---

### `comment delete`

Delete a comment.

```bash
meta-cli comment delete COMMENT_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `comment-id` | Yes | The comment ID to delete |

**Requires:** Authentication + page

---

## Insight Commands

### `insight page`

Show page-level insights (reach, impressions, engagement).

```bash
meta-cli insight page
meta-cli insight page --metric page_fans --period lifetime
meta-cli insight page --metric page_impressions,page_engaged_users --period week
meta-cli insight page --json
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metric` | string | `page_impressions,page_impressions_unique,page_engaged_users,page_post_engagements,page_views_total` | Comma-separated metrics |
| `--period` | string | `day` | Metric period (`day`, `week`, `days_28`, `month`, `lifetime`) |

**Output:** Table with metric name, period, end time, and value.

**Requires:** Authentication + page

---

### `insight post`

Show post-level insights (impressions, reach, engagement, clicks).

```bash
meta-cli insight post POST_ID
meta-cli insight post POST_ID --metric post_impressions,post_clicks
meta-cli insight post POST_ID --json
```

| Argument | Required | Description |
|----------|----------|-------------|
| `post-id` | Yes | The post ID to get insights for |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metric` | string | `post_impressions,post_impressions_unique,post_engaged_users,post_clicks` | Comma-separated metrics |

**Output:** Table with metric name, period, end time, and value.

**Requires:** Authentication + page

---

## Label Commands

### `label list`

List all custom labels for the page.

```bash
meta-cli label list
meta-cli label list --json
```

**Output:** Table with label ID and name.

**Requires:** Authentication + page

---

### `label create`

Create a new custom label.

```bash
meta-cli label create --name "VIP"
meta-cli label create -n "Support"
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `-n`, `--name` | string | Yes | Label name |

**Output:** Created label ID.

**Requires:** Authentication + page

---

### `label delete`

Delete a custom label.

```bash
meta-cli label delete LABEL_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `label-id` | Yes | The label ID to delete |

**Requires:** Authentication + page

---

### `label assign`

Assign a label to a user by their page-scoped ID (PSID).

```bash
meta-cli label assign LABEL_ID --psid USER_PSID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `label-id` | Yes | The label ID to assign |

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--psid` | string | Yes | Page-scoped user ID |

**Requires:** Authentication + page

---

### `label remove`

Remove a label from a user.

```bash
meta-cli label remove LABEL_ID --psid USER_PSID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `label-id` | Yes | The label ID to remove |

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--psid` | string | Yes | Page-scoped user ID |

**Requires:** Authentication + page

---

### `label list-by-user`

List all labels assigned to a specific user.

```bash
meta-cli label list-by-user USER_PSID
meta-cli label list-by-user USER_PSID --json
```

| Argument | Required | Description |
|----------|----------|-------------|
| `psid` | Yes | Page-scoped user ID |

**Output:** Table with label ID and name.

**Requires:** Authentication + page

---

## Messenger Commands

### `messenger send`

Send a Messenger message to a user via their page-scoped ID (PSID).

```bash
meta-cli messenger send --psid USER_PSID --message "Hello!"
meta-cli messenger send --psid USER_PSID -m "Hello!"
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--psid` | string | Yes | Page-scoped user ID |
| `-m`, `--message` | string | Yes | Message text |

**Behavior:** Sends the message via the Graph API and stores it in the local SQLite database with direction "out".

**Requires:** Authentication + page

---

### `messenger list`

List stored messages from the local SQLite database.

```bash
meta-cli messenger list
meta-cli messenger list --limit 100
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `50` | Number of messages to fetch |

**Output:** Table with message ID, PSID, text, direction (in/out), and timestamp.

**Note:** This reads from the local database only. Messages are populated by the webhook server (incoming) and the `messenger send` command (outgoing).

**Requires:** Page ID (no authentication needed - reads local DB only)

---

### `messenger history`

List conversation history with a specific user (both directions), ordered chronologically (oldest first).

```bash
meta-cli messenger history --psid USER_PSID
meta-cli messenger history --psid USER_PSID --limit 50
meta-cli messenger history --psid USER_PSID --json
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--psid` | string | (required) | Page-scoped user ID |
| `--limit` | int | `20` | Number of messages to fetch |

**Output:** Table with message ID, PSID, text, direction (in/out), and timestamp. Messages are returned in chronological order (oldest first), which is useful for viewing conversation context.

**Note:** This reads from the local database. It is designed to be called by the OpenClaw agent to get conversation context before replying.

**Requires:** Page ID (no authentication needed - reads local DB only)

---

## Config Commands

### `config set`

Set a configuration value.

```bash
meta-cli config set verify_token my-secret-token
meta-cli config set webhook_port 9090
```

| Argument | Required | Description |
|----------|----------|-------------|
| `key` | Yes | Config key to set |
| `value` | Yes | Value to assign |

**Supported keys:** `default_account`, `default_page`, `graph_api_version`, `webhook_port`, `verify_token`, `redirect_uri`, `rag_dir`, `db_path`, `debounce_seconds`, `hooks_endpoint`, `hooks_token`, `auto_reply`, `prompt_template`

**Behavior:** Loads `~/.meta-cli/config.json`, sets the field, and saves.

**Requires:** Nothing

---

### `config get`

Get a configuration value.

```bash
meta-cli config get verify_token
meta-cli config get webhook_port
```

| Argument | Required | Description |
|----------|----------|-------------|
| `key` | Yes | Config key to read |

**Output:** The value of the specified key.

**Requires:** Nothing

---

### `config list`

List all configuration values.

```bash
meta-cli config list
meta-cli config list --json
```

**Output:** Table of all config key-value pairs.

**Requires:** Nothing

---

## Webhook Commands

### `webhook serve`

Start an HTTP server to receive real-time webhook events from Meta.

```bash
# Foreground mode
meta-cli webhook serve --verify-token MY_TOKEN

# Custom port
meta-cli webhook serve --verify-token MY_TOKEN --port 9090

# Daemon mode (background)
meta-cli webhook serve --verify-token MY_TOKEN --daemon
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | from config (`8080`) | HTTP port to listen on |
| `--verify-token` | string | config or `$META_VERIFY_TOKEN` | Webhook verification token |
| `--daemon` | bool | `false` | Run as background daemon |
| `--auto-reply` | bool | `false` | Enable auto-reply via OpenClaw (overrides config) |
| `--debounce` | int | from config (`3`) | Debounce window in seconds (overrides config) |
| `--hooks-endpoint` | string | from config | OpenClaw hooks endpoint URL (overrides config) |
| `--hooks-token` | string | from config | OpenClaw hooks auth token (overrides config) |

**Behavior:**
1. Subscribes the page to webhook fields (`messages`, `message_echoes`)
2. Starts HTTP server on the specified port
3. Handles GET requests for webhook verification
4. Handles POST requests for incoming events (with HMAC-SHA256 signature validation)
5. Stores received messages in the SQLite database
6. In daemon mode: forks process, writes PID to `~/.meta-cli/webhook.pid`, logs to `~/.meta-cli/webhook.log`

**Requires:** Authentication + page

---

### `webhook subscribe`

Subscribe the page to webhook fields without starting a server.

```bash
meta-cli webhook subscribe
```

**Behavior:** Subscribes to `messages` and `message_echoes` fields via `POST /me/subscribed_apps`.

**Requires:** Authentication + page

---

### `webhook status`

Check if the webhook daemon is currently running.

```bash
meta-cli webhook status
```

**Output:** Running status and PID if active.

**Requires:** Nothing (checks local PID file)

---

### `webhook stop`

Stop the running webhook daemon.

```bash
meta-cli webhook stop
```

**Behavior:** Sends SIGTERM to the daemon process and removes the PID file. Waits up to 3 seconds for graceful shutdown.

**Requires:** Webhook daemon must be running

---

## RAG Commands

### `rag index`

Show indexing statistics for a directory of documents.

```bash
meta-cli rag index ./docs
meta-cli rag index          # uses rag_dir from config
```

| Argument | Required | Description |
|----------|----------|-------------|
| `directory` | No | Directory containing documents (defaults to `rag_dir` in config) |

**Output:** Table of indexed documents with ID, path, and title. Also shows total chunk count.

**Requires:** Nothing

---

### `rag search`

Search documents using TF-IDF ranking.

```bash
meta-cli rag search "how to reset password"
meta-cli rag search "refund policy" --top 3
meta-cli rag search "shipping" --dir ./knowledge-base
```

| Argument | Required | Description |
|----------|----------|-------------|
| `query` | Yes | Search query string |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--top` | int | `5` | Number of results to return |
| `--dir` | string | from config | Directory with documents |

**Output:** Table of search results ranked by relevance score.

**Requires:** Nothing

---

## Command Requirements Matrix

| Command | Auth | Page | API Client |
|---------|:----:|:----:|:----------:|
| `auth login` | - | - | - |
| `auth status` | - | - | - |
| `auth refresh` | Yes | - | - |
| `pages list` | Yes | - | Yes (user token) |
| `pages set-default` | - | - | - |
| `post list` | Yes | Yes | Yes |
| `post create` | Yes | Yes | Yes |
| `post update` | Yes | Yes | Yes |
| `post edit` | Yes | Yes | Yes |
| `post delete` | Yes | Yes | Yes |
| `post list-scheduled` | Yes | Yes | Yes |
| `comment list` | Yes | Yes | Yes |
| `comment reply` | Yes | Yes | Yes |
| `comment update` | Yes | Yes | Yes |
| `comment hide` | Yes | Yes | Yes |
| `comment unhide` | Yes | Yes | Yes |
| `comment delete` | Yes | Yes | Yes |
| `insight page` | Yes | Yes | Yes |
| `insight post` | Yes | Yes | Yes |
| `label list` | Yes | Yes | Yes |
| `label create` | Yes | Yes | Yes |
| `label delete` | Yes | Yes | Yes |
| `label assign` | Yes | Yes | Yes |
| `label remove` | Yes | Yes | Yes |
| `label list-by-user` | Yes | Yes | Yes |
| `messenger send` | Yes | Yes | Yes |
| `messenger list` | - | Yes | - |
| `messenger history` | - | Yes | - |
| `webhook serve` | Yes | Yes | Yes |
| `webhook subscribe` | Yes | Yes | Yes |
| `webhook status` | - | - | - |
| `webhook stop` | - | - | - |
| `config set` | - | - | - |
| `config get` | - | - | - |
| `config list` | - | - | - |
| `rag index` | - | - | - |
| `rag search` | - | - | - |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `META_VERIFY_TOKEN` | Webhook verification token (alternative to `--verify-token` flag or `config set verify_token`) |
