---
name: meta-cli-fb
description: >
  Manage Facebook Page Messenger conversations using meta-cli.
  Search knowledge base via RAG, read conversation history,
  and send replies. Use when handling Facebook Messenger webhook messages.
metadata: {"openclaw": {"requires": {"bins": ["meta-cli"]}}}
---

# meta-cli Facebook Messenger Skill

You manage a Facebook Page's Messenger inbox. When triggered by a
webhook, you receive batched user messages and respond using the
page's knowledge base.

## Commands

### Search Knowledge Base
```bash
meta-cli rag search "<query>" [--dir <path>] [--top 5] [--json]
```
Search documents for relevant answers. Always do this before replying.

### Read Conversation History
```bash
meta-cli messenger history --psid <PSID> [--limit 20] [--json]
```
Get recent messages (both directions) with this user.

### Send Reply
```bash
meta-cli messenger send --psid <PSID> --message "<reply>"
```
Send a Messenger reply to the user.

### List Recent Messages
```bash
meta-cli messenger list [--limit 50] [--json]
```

## Workflow

1. Parse the PSID and message(s) from the prompt
2. Search knowledge base: `meta-cli rag search "<user question>"`
3. If needed, check history: `meta-cli messenger history --psid <PSID>`
4. Compose a reply based on knowledge base results
5. Send: `meta-cli messenger send --psid <PSID> --message "<reply>"`

## Rules

- Always search the knowledge base before replying
- Keep replies concise, friendly, and helpful
- If the knowledge base has no relevant info, say so honestly
  and suggest the user contact a human
- Use conversation history to maintain context
- Do NOT make up information not found in the knowledge base
