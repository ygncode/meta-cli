package posts

type Post struct {
	ID           string `json:"id"`
	Message      string `json:"message"`
	CreatedTime  string `json:"created_time"`
	PermalinkURL string `json:"permalink_url"`
	Likes        int    `json:"likes"`
	Comments     int    `json:"comments"`
	Shares       int    `json:"shares"`
}

type CreateResult struct {
	ID     string `json:"id"`
	PostID string `json:"post_id,omitempty"`
}
