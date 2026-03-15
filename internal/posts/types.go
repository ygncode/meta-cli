package posts

import "time"

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

type ScheduleOpts struct {
	PublishTime time.Time
}

type VideoOpts struct {
	FilePath  string
	Title     string
	Message   string
	Thumbnail string
}

type ScheduledPost struct {
	ID                   string `json:"id"`
	Message              string `json:"message"`
	ScheduledPublishTime string `json:"scheduled_publish_time"`
	CreatedTime          string `json:"created_time"`
}
