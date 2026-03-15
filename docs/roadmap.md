# Feature Roadmap

Graph API capabilities not yet covered by meta-cli, prioritized by impact.

## Tier 1 — High Impact

### ~~Page & Post Insights~~ ✅ Completed
Implemented as `insight page` and `insight post` commands with `--metric` and `--period` flags.

### Video Upload
Video is Facebook's dominant content type. The CLI supports photos but not video.
- `POST /{page-id}/videos` with source, title, description, thumbnail
- Supports `scheduled_publish_time` for scheduled video posts

### Reels Publishing
Facebook's fastest-growing format. Limit: 30 per 24h.
- `POST /{page-id}/video_reels` — 3-step process: init upload, upload video, finish/publish
- Supports description, scheduled publishing

### Page Info Display
Read and display page metadata from the terminal.
- `GET /{page-id}?fields=about,description,hours,phone,website,email,location,category,fan_count,followers_count,verification_status,...`
- Dozens of readable fields currently not surfaced

### Messenger Attachments
Send images, video, audio, and files (up to 25MB), not just text.
- Send API with `attachment` payload instead of `message.text`
- Supports image, video, audio, and file types

## Tier 2 — Medium Impact

### Reactions Listing
Breakdown by reaction type (LIKE, LOVE, WOW, HAHA, SAD, ANGRY, CARE) on posts and comments.
- `GET /{post-id}/reactions?summary=true`
- `GET /{comment-id}/reactions`

### Blocked User Management
Moderation workflow for blocking/unblocking users.
- `GET /{page-id}/blocked` — list blocked users
- `POST /{page-id}/blocked` — block a user (accepts user IDs, ASIDs, PSIDs)
- `DELETE /{page-id}/blocked` — unblock a user

### Visitor & Tagged Posts
Monitor user-generated content and brand mentions.
- `GET /{page-id}/visitor_posts` — posts by others on your page
- `GET /{page-id}/tagged` — posts where your page is tagged

### Messenger Templates & Quick Replies
Structured interactive messages beyond plain text.
- Generic, button, media, product, receipt, coupon templates via Send API
- Quick replies: up to 13 buttons, including email/phone collection

### Private Reply to Comment
Move public discussions to private Messenger conversation.
- Send API with `recipient: {comment_id: ...}`

### Message Tags
Send messages outside the standard 24-hour messaging window.
- Tags: `HUMAN_AGENT`, `ACCOUNT_UPDATE`, `POST_PURCHASE_UPDATE`, `CONFIRMED_EVENT_UPDATE`
- Send API with `messaging_type: MESSAGE_TAG`

### User / Role Management
Manage who has access to the page and their roles.
- `GET /{page-id}/assigned_users` — list users with access
- `POST /{page-id}/assigned_users` — assign roles (MANAGE, CREATE_CONTENT, MODERATE, MESSAGING, ADVERTISE, ANALYZE)
- `DELETE /{page-id}/assigned_users?user={id}` — remove access

### Ratings / Reviews
Monitor page recommendations and reputation.
- `GET /{page-id}/ratings` — list reviews (read-only)
- `GET /{page-id}?fields=overall_star_rating,rating_count` — overall rating

### Post CTA Buttons
Add call-to-action buttons to posts.
- `call_to_action` parameter on feed post creation
- Types: SHOP_NOW, LEARN_MORE, SIGN_UP, BOOK_TRAVEL, DOWNLOAD, etc.

## Tier 3 — Lower Priority

### Messenger Profile Configuration
Configure the Messenger experience for your page.
- `POST /me/messenger_profile` with `persistent_menu` — always-visible menu
- `POST /me/messenger_profile` with `get_started` — first interaction payload
- `POST /me/messenger_profile` with `greeting` — welcome screen text
- `POST /me/messenger_profile` with `ice_breakers` — suggested conversation starters

### Conversations List from API
List conversations directly from the Graph API (currently we only list from local SQLite).
- `GET /{page-id}/conversations?platform=MESSENGER`

### Lead Generation
Create lead capture forms and retrieve collected leads.
- `POST /{page-id}/leadgen_forms` — create forms
- `GET /{leadgen-form-id}/leads` — retrieve leads

### Backdated Posts & Audience Targeting
- `backdated_time` + `backdated_time_granularity` params — backdate historical content
- `targeting` and `feed_targeting` params — restrict visibility by geo/demographics
- `place` param — tag posts with a location

### Events
Read-only access to page events.
- `GET /{page-id}/events` — list events
- Cannot create/update/delete events via API
