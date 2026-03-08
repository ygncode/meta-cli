---
name: meta-cli-fb
description: >
  Manage a Facebook Page using meta-cli. Handles Messenger conversations
  (search knowledge base, read history, send replies), page posts
  (list, create, delete), comments (list, reply, hide, unhide, delete),
  RAG-powered knowledge base search, and page administration.
  Use when handling Facebook Page tasks triggered by webhooks or user requests.
metadata: {"openclaw": {"requires": {"bins": ["meta-cli"]}}}
---

# meta-cli Facebook Page Skill

You manage a Facebook Page end-to-end using `meta-cli`. This includes
responding to Messenger conversations, managing posts and comments,
and searching the knowledge base. All output supports `--json` for
structured parsing.

## Messenger Commands

### Send Reply
```bash
meta-cli messenger send --psid <PSID> --message "<reply>"
```
Send a Messenger reply to a user.

### Read Conversation History
```bash
meta-cli messenger history --psid <PSID> [--limit 20] [--json]
```
Get recent messages (both in and out) with a user, oldest first.

### List All Messages
```bash
meta-cli messenger list [--limit 50] [--json]
```
List all stored messages across all users.

## Post Commands

### List Posts
```bash
meta-cli post list [--limit 10] [--json]
```
List recent page posts with ID, message, creation time, permalink, and engagement (likes, comments, shares).

### Create Post
```bash
# Text post
meta-cli post create --message "Hello world!"

# Photo post
meta-cli post create --photo /path/to/image.jpg --message "Check this out!"

# Multi-photo album
meta-cli post create --photo img1.jpg --photo img2.jpg --message "Album"

# Link post
meta-cli post create --link https://example.com --message "Read this"
```
At least one of `--message`, `--photo`, or `--link` is required.

### Delete Post
```bash
meta-cli post delete <POST_ID>
```

## Comment Commands

### List Comments
```bash
meta-cli comment list <POST_ID> [--limit 25] [--json]
```
List comments on a post with author, message, timestamp, and like count.

### Reply to Comment
```bash
meta-cli comment reply <COMMENT_ID> --message "<reply>"
```

### Hide / Unhide Comment
```bash
meta-cli comment hide <COMMENT_ID>
meta-cli comment unhide <COMMENT_ID>
```

### Delete Comment
```bash
meta-cli comment delete <COMMENT_ID>
```

## RAG (Knowledge Base) Commands

### Search Knowledge Base
```bash
meta-cli rag search "<query>" [--dir <path>] [--top 5] [--json]
```
Search documents for relevant answers using TF-IDF ranking. Always do this before replying to user questions.

### Index Stats
```bash
meta-cli rag index [<directory>] [--json]
```
Show indexed documents and chunk count.

## Page & Auth Commands

### List Pages
```bash
meta-cli pages list [--json]
```

### Check Auth Status
```bash
meta-cli auth status [--json]
```

### View / Set Config
```bash
meta-cli config list [--json]
meta-cli config get <key>
meta-cli config set <key> <value>
```

## Global Flags

All commands support these flags:
- `--json` — Output as JSON (useful for parsing results)
- `--plain` — Output as TSV
- `--page <PAGE_ID>` — Override the default page

## Messenger Workflow

When triggered by a webhook with incoming user messages:

1. Parse the PSID and message(s) from the prompt
2. Search knowledge base: `meta-cli rag search "<user question>"`
3. If needed, check history: `meta-cli messenger history --psid <PSID>`
4. Compose a reply based on knowledge base results
5. Send: `meta-cli messenger send --psid <PSID> --message "<reply>"`

## Comment Management Workflow

When asked to moderate or respond to comments:

1. List recent posts: `meta-cli post list --json`
2. List comments on a post: `meta-cli comment list <POST_ID> --json`
3. Take action: reply, hide inappropriate comments, or delete spam
4. If replying, search knowledge base first for accurate answers

## Rules

- Always search the knowledge base before replying to questions
- Keep replies concise, friendly, and helpful
- If the knowledge base has no relevant info, say so honestly and suggest contacting a human
- Use conversation history to maintain context in Messenger threads
- Do NOT make up information not found in the knowledge base
- When moderating comments, hide rather than delete unless it's clear spam
- Use `--json` flag when you need to parse output programmatically
