# Commands Reference

## Command Tree

```
meta-cli
├── auth
│   ├── login              Login with Facebook OAuth
│   ├── status             Show current auth status
│   └── refresh            Refresh page tokens
├── pages
│   ├── list               List pages you manage
│   ├── set-default        Set a default page for commands
│   └── info               Display page information
├── post (alias: posts)
│   ├── list               List recent posts
│   ├── create             Create a text, photo, video, or link post
│   ├── update             Update a post's message
│   ├── edit               Edit a post's message
│   ├── delete             Delete a post
│   ├── list-scheduled     List scheduled (unpublished) posts
│   ├── list-visitor       List visitor posts on the page
│   └── list-tagged        List posts where the page is tagged
├── reel (alias: reels)
│   └── create             Publish a reel (short-form video)
├── comment (alias: comments)
│   ├── list               List comments on a post
│   ├── reply              Reply to a comment
│   ├── update             Update a comment's message
│   ├── hide               Hide a comment
│   ├── unhide             Unhide a comment
│   ├── delete             Delete a comment
│   └── private-reply      Send a private Messenger reply
├── insight (alias: insights)
│   ├── page               Show page-level insights
│   └── post               Show post-level insights
├── label (alias: labels)
│   ├── list               List all custom labels for the page
│   ├── create             Create a new custom label
│   ├── delete             Delete a custom label
│   ├── assign             Assign a label to a user
│   ├── remove             Remove a label from a user
│   └── list-by-user       List labels assigned to a user
├── messenger
│   ├── send               Send a message (text, attachment, tagged, quick replies)
│   ├── send-template      Send a template message
│   ├── list               List stored messages
│   ├── history            List conversation history with a user
│   ├── conversations      List Messenger conversations from API
│   └── profile
│       ├── get            Get Messenger profile settings
│       ├── set-greeting   Set the greeting text
│       ├── set-get-started Set the Get Started button
│       ├── set-menu       Set the persistent menu
│       ├── set-ice-breakers Set ice breaker starters
│       └── delete         Delete a profile field
├── event (alias: events)
│   └── list               List page events
├── rating (alias: ratings)
│   ├── list               List page ratings and reviews
│   └── summary            Show overall page rating
├── reaction (alias: reactions)
│   └── list               List reactions on a post or comment
├── blocked
│   ├── list               List blocked users
│   ├── add                Block a user
│   └── remove             Unblock a user
├── role (alias: roles)
│   ├── list               List users with page access
│   ├── assign             Assign roles to a user
│   └── remove             Remove a user's page access
├── lead (alias: leads)
│   ├── create-form        Create a lead generation form
│   └── list               List leads from a form
├── config
│   ├── set                Set a config value
│   ├── get                Get a config value
│   └── list               List all config values
├── webhook
│   ├── serve              Start webhook HTTP server
│   ├── subscribe          Subscribe page to webhook fields
│   ├── status             Check if webhook server is running
│   └── stop               Stop the webhook server
└── rag
    ├── index              Show index stats for documents
    └── search             Search documents by query
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

### `pages info`

Display page information including metadata, contact info, and follower stats.

```bash
meta-cli pages info
meta-cli pages info --json
```

**Output:** Page ID, name, about, description, category, phone, website, emails, fan count, followers count, and verification status.

**Requires:** Authentication + page

---

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
| `--backdate` | string | No | Backdate post (format: `"YYYY-MM-DD"`) |
| `--backdate-granularity` | string | No | Backdate granularity (`year`, `month`, `day`, `hour`, `min`) |
| `--targeting` | string | No | Audience targeting as JSON (e.g. `'{"geo_locations":{"countries":["US"]}}'`) |
| `--place` | string | No | Place ID to tag the post with a location |
| `--cta` | string | No | Call-to-action as JSON (e.g. `'{"type":"SHOP_NOW","value":{"link":"https://..."}}'`) |

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

### `post list-visitor`

List posts by visitors on the page.

```bash
meta-cli post list-visitor
meta-cli post list-visitor --limit 50
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `25` | Number of visitor posts to fetch |

**Output:** Table with post ID, message, author, and creation time.

**Requires:** Authentication + page

---

### `post list-tagged`

List posts where the page is tagged.

```bash
meta-cli post list-tagged
meta-cli post list-tagged --limit 50
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `25` | Number of tagged posts to fetch |

**Output:** Table with post ID, message, author, and creation time.

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

## Reel Commands

### `reel create`

Publish a reel (short-form video) on the page. Uses the 3-step Reels API: init upload session, upload binary, finish/publish.

```bash
# Basic reel
meta-cli reel create --video /path/to/clip.mp4 --message "Check this out!"

# Reel with title
meta-cli reel create --video clip.mp4 --title "My Reel" --message "Description"

# Scheduled reel
meta-cli reel create --video clip.mp4 --message "Coming soon!" --schedule "2026-03-20 14:00"
meta-cli reel create --video clip.mp4 --message "Hello!" --schedule "2026-03-20 14:00" --tz "Asia/Yangon"
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--video` | string | Yes | Path to video file |
| `--message` | string | No | Reel description |
| `--title` | string | No | Reel title |
| `--schedule` | string | No | Schedule for future publishing (format: `"YYYY-MM-DD HH:MM"`) |
| `--tz` | string | No | Timezone for `--schedule` (e.g. `"Asia/Yangon"`), defaults to local |

**Output:** Created reel details (ID, post ID).

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

### `comment private-reply`

Send a private Messenger reply to a comment.

```bash
meta-cli comment private-reply COMMENT_ID --message "Let's discuss this privately"
meta-cli comment private-reply COMMENT_ID -m "I'll help you via DM"
```

| Argument | Required | Description |
|----------|----------|-------------|
| `comment-id` | Yes | The comment ID to reply to privately |

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `-m`, `--message` | string | Yes | Message text |

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
| `-m`, `--message` | string | No* | Message text |
| `--image` | string | No* | Image URL or local file path |
| `--video` | string | No* | Video URL or local file path |
| `--audio` | string | No* | Audio URL or local file path |
| `--file` | string | No* | File URL or local file path |
| `--tag` | string | No | Message tag for outside 24-hour window (HUMAN_AGENT, ACCOUNT_UPDATE, POST_PURCHASE_UPDATE, CONFIRMED_EVENT_UPDATE) |
| `--quick-reply` | string[] | No | Quick reply option (repeatable) |

*At least one of `--message` or an attachment flag is required.

**Behavior:** Sends the message via the Graph API and stores it in the local SQLite database with direction "out". If a URL or file attachment is provided, it is sent as a media attachment. Quick replies add interactive buttons below the message.

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

### `messenger send-template`

Send a structured template message.

```bash
meta-cli messenger send-template --psid USER_PSID --json '{"template_type":"button","text":"Hello","buttons":[]}'
meta-cli messenger send-template --psid USER_PSID --file template.json
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--psid` | string | Yes | Page-scoped user ID |
| `--json` | string | No* | Template payload as JSON string |
| `--file` | string | No* | Path to JSON file with template payload |

*One of `--json` or `--file` is required.

**Requires:** Authentication + page

---

### `messenger conversations`

List Messenger conversations from the API.

```bash
meta-cli messenger conversations
meta-cli messenger conversations --limit 50
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `25` | Number of conversations to fetch |

**Output:** Table with conversation ID, participants, updated time, and message count.

**Requires:** Authentication + page

---

### `messenger profile get`

Get current Messenger profile settings.

```bash
meta-cli messenger profile get
```

**Requires:** Authentication + page

---

### `messenger profile set-greeting`

Set the Messenger greeting text.

```bash
meta-cli messenger profile set-greeting "Welcome to our page!"
```

**Requires:** Authentication + page

---

### `messenger profile set-get-started`

Set the Get Started button payload.

```bash
meta-cli messenger profile set-get-started GET_STARTED_PAYLOAD
```

**Requires:** Authentication + page

---

### `messenger profile set-menu`

Set the persistent menu.

```bash
meta-cli messenger profile set-menu --json '[{"locale":"default","call_to_actions":[...]}]'
meta-cli messenger profile set-menu --file menu.json
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--json` | string | No* | Menu definition as JSON |
| `--file` | string | No* | Path to JSON file |

**Requires:** Authentication + page

---

### `messenger profile set-ice-breakers`

Set ice breaker conversation starters.

```bash
meta-cli messenger profile set-ice-breakers --json '[{"question":"Help?","payload":"HELP"}]'
meta-cli messenger profile set-ice-breakers --file icebreakers.json
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--json` | string | No* | Ice breakers as JSON |
| `--file` | string | No* | Path to JSON file |

**Requires:** Authentication + page

---

### `messenger profile delete`

Delete a Messenger profile field.

```bash
meta-cli messenger profile delete --field greeting
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--field` | string | Yes | Profile field to delete |

**Requires:** Authentication + page

---

## Event Commands

### `event list`

List page events.

```bash
meta-cli event list
meta-cli event list --limit 50
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `25` | Number of events to fetch |

**Output:** Table with event ID, name, description, start time, end time, and place.

**Requires:** Authentication + page

---

## Rating Commands

### `rating list`

List page ratings and reviews.

```bash
meta-cli rating list
meta-cli rating list --limit 50
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `25` | Number of ratings to fetch |

**Output:** Table with reviewer name, rating, review text, and creation time.

**Requires:** Authentication + page

---

### `rating summary`

Show overall page rating.

```bash
meta-cli rating summary
meta-cli rating summary --json
```

**Output:** Star rating and total rating count.

**Requires:** Authentication + page

---

## Reaction Commands

### `reaction list`

List reactions on a post or comment.

```bash
meta-cli reaction list POST_ID
meta-cli reaction list COMMENT_ID --limit 100
```

| Argument | Required | Description |
|----------|----------|-------------|
| `object-id` | Yes | Post or comment ID |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `50` | Number of reactions to fetch |

**Output:** Table with user ID, name, and reaction type (LIKE, LOVE, WOW, HAHA, SAD, ANGRY, CARE).

**Requires:** Authentication + page

---

## Blocked User Commands

### `blocked list`

List blocked users.

```bash
meta-cli blocked list
meta-cli blocked list --limit 50
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `25` | Number of blocked users to fetch |

**Requires:** Authentication + page

---

### `blocked add`

Block a user.

```bash
meta-cli blocked add USER_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `user-id` | Yes | User ID to block |

**Requires:** Authentication + page

---

### `blocked remove`

Unblock a user.

```bash
meta-cli blocked remove USER_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `user-id` | Yes | User ID to unblock |

**Requires:** Authentication + page

---

## Role Commands

### `role list`

List users with page access.

```bash
meta-cli role list
meta-cli role list --json
```

**Output:** Table with user ID, name, and assigned tasks.

**Requires:** Authentication + page

---

### `role assign`

Assign roles to a user.

```bash
meta-cli role assign USER_ID --tasks MANAGE,CREATE_CONTENT
```

| Argument | Required | Description |
|----------|----------|-------------|
| `user-id` | Yes | User ID to assign roles to |

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--tasks` | string | Yes | Comma-separated tasks (MANAGE, CREATE_CONTENT, MODERATE, MESSAGING, ADVERTISE, ANALYZE) |

**Requires:** Authentication + page

---

### `role remove`

Remove a user's page access.

```bash
meta-cli role remove USER_ID
```

| Argument | Required | Description |
|----------|----------|-------------|
| `user-id` | Yes | User ID to remove |

**Requires:** Authentication + page

---

## Lead Commands

### `lead create-form`

Create a lead generation form.

```bash
meta-cli lead create-form --json '{"name":"Contact Form","questions":[{"type":"FULL_NAME"}]}'
meta-cli lead create-form --file form.json
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--json` | string | No* | Form definition as JSON |
| `--file` | string | No* | Path to JSON file |

**Requires:** Authentication + page

---

### `lead list`

List leads from a form.

```bash
meta-cli lead list FORM_ID
meta-cli lead list FORM_ID --limit 100
```

| Argument | Required | Description |
|----------|----------|-------------|
| `form-id` | Yes | Lead gen form ID |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `50` | Number of leads to fetch |

**Requires:** Authentication + page

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
| `pages info` | Yes | Yes | Yes |
| `post list` | Yes | Yes | Yes |
| `post create` | Yes | Yes | Yes |
| `post update` | Yes | Yes | Yes |
| `post edit` | Yes | Yes | Yes |
| `post delete` | Yes | Yes | Yes |
| `post list-scheduled` | Yes | Yes | Yes |
| `post list-visitor` | Yes | Yes | Yes |
| `post list-tagged` | Yes | Yes | Yes |
| `reel create` | Yes | Yes | Yes |
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
| `comment private-reply` | Yes | Yes | Yes |
| `messenger send` | Yes | Yes | Yes |
| `messenger send-template` | Yes | Yes | Yes |
| `messenger list` | - | Yes | - |
| `messenger history` | - | Yes | - |
| `messenger conversations` | Yes | Yes | Yes |
| `messenger profile get` | Yes | Yes | Yes |
| `messenger profile set-*` | Yes | Yes | Yes |
| `messenger profile delete` | Yes | Yes | Yes |
| `event list` | Yes | Yes | Yes |
| `rating list` | Yes | Yes | Yes |
| `rating summary` | Yes | Yes | Yes |
| `reaction list` | Yes | Yes | Yes |
| `blocked list` | Yes | Yes | Yes |
| `blocked add` | Yes | Yes | Yes |
| `blocked remove` | Yes | Yes | Yes |
| `role list` | Yes | Yes | Yes |
| `role assign` | Yes | Yes | Yes |
| `role remove` | Yes | Yes | Yes |
| `lead create-form` | Yes | Yes | Yes |
| `lead list` | Yes | Yes | Yes |
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
