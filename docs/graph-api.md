# Graph API Client

## Overview

The `internal/graph` package provides an HTTP client for the Meta Graph API. It handles request construction, authentication, response parsing, and error handling.

## Client

**File:** `internal/graph/client.go`

### Initialization

```go
// Create a new client with API version and access token
client := graph.New("v25.0", pageAccessToken)

// Create with custom HTTP client (for testing)
client := graph.NewWithHTTPClient(baseURL, token, httpClient)

// Create a new client with a different token (same base URL)
newClient := client.WithToken(differentToken)
```

### Client Structure

```go
type Client struct {
    baseURL    string       // "https://graph.facebook.com/v25.0"
    httpClient *http.Client // HTTP client (defaults to http.DefaultClient)
    token      string       // Access token appended to all requests
}
```

## HTTP Methods

### GET

```go
func (c *Client) Get(ctx context.Context, path string, params url.Values, out any) error
```

Sends a GET request with URL-encoded query parameters. The access token is automatically added.

**Usage in services:**
```go
// List posts
client.Get(ctx, pageID+"/posts", url.Values{
    "fields": {"id,message,created_time,permalink_url,likes.summary(true),comments.summary(true),shares"},
    "limit":  {strconv.Itoa(limit)},
}, &response)

// List page accounts
client.Get(ctx, "me/accounts", url.Values{
    "fields": {"id,name,access_token"},
}, &response)
```

### POST (Form-encoded)

```go
func (c *Client) Post(ctx context.Context, path string, body url.Values, out any) error
```

Sends a POST request with `application/x-www-form-urlencoded` body.

**Usage in services:**
```go
// Create a text post
client.Post(ctx, pageID+"/feed", url.Values{
    "message": {message},
}, &result)

// Subscribe to webhook fields
client.Post(ctx, "me/subscribed_apps", url.Values{
    "subscribed_fields": {"messages,message_echoes"},
}, &result)
```

### POST (Multipart)

```go
func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string, filePath string, out any) error
```

Sends a multipart form request for file uploads. Uses `io.Pipe` for streaming the file without loading it entirely into memory.

**Usage in services:**
```go
// Upload a photo
client.PostMultipart(ctx, pageID+"/photos", map[string]string{
    "message":   message,
    "published": "true",
}, filePath, &result)
```

**Implementation details:**
- Uses `io.Pipe()` for streaming multipart encoding
- File is written to the multipart writer in a goroutine
- The file field name is `"source"` with the original filename

### DELETE

```go
func (c *Client) Delete(ctx context.Context, path string, out any) error
```

Sends a DELETE request.

**Usage in services:**
```go
// Delete a post
client.Delete(ctx, postID, &result)
```

### DELETE with Parameters

```go
func (c *Client) DeleteWithParams(ctx context.Context, path string, params url.Values, out any) error
```

Sends a DELETE request with URL-encoded query parameters. Used when the API requires additional data on a DELETE request (e.g., removing a label from a user).

**Usage in services:**
```go
// Remove a label from a user
client.DeleteWithParams(ctx, labelID+"/label", url.Values{"user": {userPSID}}, &result)
```

## Request Pipeline

All methods go through a common `do()` function:

```
Client.Get/Post/PostMultipart/Delete/DeleteWithParams
  └── Build *http.Request with context
       └── do(req, out)
            ├── httpClient.Do(req)           # Execute HTTP request
            ├── io.ReadAll(resp.Body)         # Read response body
            ├── if status >= 400:
            │     ├── Try parse GraphError envelope
            │     └── Return *APIError
            └── if status < 400:
                  └── json.Unmarshal(body, out) # Parse response into out
```

## Error Handling

**File:** `internal/graph/errors.go`

### Error Types

```go
type GraphError struct {
    Message  string `json:"message"`
    Type     string `json:"type"`
    Code     int    `json:"code"`
    FBTraceID string `json:"fbtrace_id"`
}

type APIError struct {
    StatusCode int
    Graph      *GraphError
}
```

### Error Response Format

The Meta Graph API returns errors in this envelope:

```json
{
  "error": {
    "message": "Invalid OAuth access token.",
    "type": "OAuthException",
    "code": 190,
    "fbtrace_id": "AbCdEfGhIjK"
  }
}
```

### Error Helpers

```go
// Check if the error is a token expiration (code 190)
func IsTokenExpired(err error) bool

// Check if the error is a permission denial (code 200 or 10)
func IsPermissionDenied(err error) bool
```

### Common Error Codes

| Code | Type | Meaning |
|------|------|---------|
| 190 | `OAuthException` | Access token expired or invalid |
| 200 | `OAuthException` | Insufficient permissions |
| 10 | `OAuthException` | Application does not have permission |
| 100 | `GraphMethodException` | Invalid parameter |
| 803 | `OAuthException` | Invalid ID |

## API Endpoints Used

### Pages

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List pages | GET | `/me/accounts` |

### Posts

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List posts | GET | `/{page_id}/posts` |
| Create text post | POST | `/{page_id}/feed` |
| Create photo post | POST (multipart) | `/{page_id}/photos` |
| Upload unpublished photo | POST (multipart) | `/{page_id}/photos` (published=false) |
| Create album post | POST | `/{page_id}/feed` (with attached_media) |
| Create link post | POST | `/{page_id}/feed` (with link) |
| Update post message | POST | `/{post_id}` (with message) |
| Delete post | DELETE | `/{post_id}` |

### Comments

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List comments | GET | `/{post_id}/comments` |
| Reply to comment | POST | `/{comment_id}/comments` |
| Update comment message | POST | `/{comment_id}` (with message) |
| Hide/unhide comment | POST | `/{comment_id}` (is_hidden=true/false) |
| Delete comment | DELETE | `/{comment_id}` |

### Labels

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List labels | GET | `/{page_id}/custom_labels` |
| Create label | POST | `/{page_id}/custom_labels` |
| Delete label | DELETE | `/{label_id}` |
| Assign label to user | POST | `/{label_id}/label` (body: user=PSID) |
| Remove label from user | DELETE | `/{label_id}/label` (query: user=PSID) |
| List labels by user | GET | `/{user_psid}/custom_labels` |

### Messenger

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Send message | POST | `/me/messages` |
| Subscribe webhook | POST | `/me/subscribed_apps` |

### Authentication (direct HTTP, not via Client)

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Exchange code for token | GET | `/oauth/access_token` |
| Extend token | GET | `/oauth/access_token` (grant_type=fb_exchange_token) |
| Fetch page tokens | GET | `/me/accounts` |

## Token Management

The client is designed to be used with page-specific tokens. When a command requires API access:

```
requirePageClient(cmd)
  ├── Load tokens from keyring
  ├── Find page token for the active page ID
  └── Create graph.Client with that page token
```

The `WithToken()` method creates a new client with a different token, which is used when switching between user-level and page-level tokens (e.g., `pages list` uses the user token, while `post list` uses the page token).
