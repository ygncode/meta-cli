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
│   ├── create         Create a text, photo, or link post
│   └── delete         Delete a post
├── comment (alias: comments)
│   ├── list           List comments on a post
│   ├── reply          Reply to a comment
│   ├── hide           Hide a comment
│   ├── unhide         Unhide a comment
│   └── delete         Delete a comment
├── messenger
│   ├── send           Send a Messenger message
│   └── list           List stored messages
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

Create a new post on the page. Supports text, photo, multi-photo album, and link posts.

```bash
# Text post
meta-cli post create --message "Hello world!"

# Single photo post
meta-cli post create --photo /path/to/image.jpg --message "Check this out!"

# Multi-photo album post
meta-cli post create --photo img1.jpg --photo img2.jpg --message "Album post"

# Link post
meta-cli post create --link https://example.com --message "Interesting link"
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--message` | string | No* | Post text content |
| `--photo` | string[] | No* | Path to image file (repeatable for albums) |
| `--link` | string | No* | URL to share |

*At least one of `--message`, `--photo`, or `--link` is required.

**Post type resolution:**
1. If multiple `--photo` flags → Album post (photos uploaded as unpublished, then attached)
2. If single `--photo` → Photo post (multipart upload)
3. If `--link` (no photos) → Link post
4. If only `--message` → Text post

**Output:** Created post details (ID, permalink).

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
| `--verify-token` | string | `$META_VERIFY_TOKEN` | Webhook verification token |
| `--daemon` | bool | `false` | Run as background daemon |

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
| `post delete` | Yes | Yes | Yes |
| `comment list` | Yes | Yes | Yes |
| `comment reply` | Yes | Yes | Yes |
| `comment hide` | Yes | Yes | Yes |
| `comment unhide` | Yes | Yes | Yes |
| `comment delete` | Yes | Yes | Yes |
| `messenger send` | Yes | Yes | Yes |
| `messenger list` | - | Yes | - |
| `webhook serve` | Yes | Yes | Yes |
| `webhook subscribe` | Yes | Yes | Yes |
| `webhook status` | - | - | - |
| `webhook stop` | - | - | - |
| `rag index` | - | - | - |
| `rag search` | - | - | - |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `META_VERIFY_TOKEN` | Webhook verification token (alternative to `--verify-token` flag) |
