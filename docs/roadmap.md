# Feature Roadmap

Graph API capabilities not yet covered by meta-cli, prioritized by impact.

## Tier 1 — High Impact

### ~~Page & Post Insights~~ ✅ Completed
Implemented as `insight page` and `insight post` commands with `--metric` and `--period` flags.

### ~~Video Upload~~ ✅ Completed
Implemented as `--video`, `--title`, and `--thumbnail` flags on `post create`. Supports scheduling with `--schedule` and `--tz`.

### ~~Reels Publishing~~ ✅ Completed
Implemented as `reel create` command with `--video`, `--message`, `--title`, `--schedule`, and `--tz` flags. Uses 3-step upload process (init, binary upload, finish/publish).

### ~~Page Info Display~~ ✅ Completed
Implemented as `pages info` command. Displays page metadata including name, about, category, phone, website, emails, fan count, followers count, and verification status.

### ~~Messenger Attachments~~ ✅ Completed
Implemented as `--image`, `--video`, `--audio`, and `--file` flags on `messenger send`. Supports both URL-based and local file upload. Stores descriptive text in local database.

## Tier 2 — Medium Impact

### ~~Reactions Listing~~ ✅ Completed
Implemented as `reaction list OBJECT_ID` command. Works for both post and comment IDs with `--limit` flag.

### ~~Blocked User Management~~ ✅ Completed
Implemented as `blocked list`, `blocked add USER_ID`, and `blocked remove USER_ID` commands.

### ~~Visitor & Tagged Posts~~ ✅ Completed
Implemented as `post list-visitor` and `post list-tagged` commands with `--limit` flag.

### ~~Messenger Templates & Quick Replies~~ ✅ Completed
Implemented as `messenger send-template --psid USER --json '{...}'` and `--quick-reply` (repeatable) flag on `messenger send`.

### ~~Private Reply to Comment~~ ✅ Completed
Implemented as `comment private-reply COMMENT_ID --message "..."` command.

### ~~Message Tags~~ ✅ Completed
Implemented as `--tag` flag on `messenger send`. Supports HUMAN_AGENT, ACCOUNT_UPDATE, POST_PURCHASE_UPDATE, CONFIRMED_EVENT_UPDATE.

### ~~User / Role Management~~ ✅ Completed
Implemented as `role list`, `role assign USER_ID --tasks MANAGE,CREATE_CONTENT`, and `role remove USER_ID` commands.

### ~~Ratings / Reviews~~ ✅ Completed
Implemented as `rating list` and `rating summary` commands with `--limit` flag.

### ~~Post CTA Buttons~~ ✅ Completed
Implemented as `--cta` flag on `post create`. Accepts JSON with CTA type and value (e.g. SHOP_NOW, LEARN_MORE, SIGN_UP, BOOK_TRAVEL, DOWNLOAD).

## Tier 3 — Lower Priority

### ~~Messenger Profile Configuration~~ ✅ Completed
Implemented as `messenger profile` subcommands: `get`, `set-greeting`, `set-get-started`, `set-menu`, `set-ice-breakers`, and `delete`.

### ~~Conversations List from API~~ ✅ Completed
Implemented as `messenger conversations` command with `--limit` flag.

### ~~Lead Generation~~ ✅ Completed
Implemented as `lead create-form --json '{...}'` and `lead list FORM_ID` commands.

### ~~Backdated Posts & Audience Targeting~~ ✅ Completed
Implemented as `--backdate`, `--backdate-granularity`, `--targeting`, and `--place` flags on `post create`.

### ~~Events~~ ✅ Completed
Implemented as `event list` command with `--limit` flag. Read-only access to page events.
