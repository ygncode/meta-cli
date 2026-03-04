package comments

type Comment struct {
	ID          string `json:"id"`
	Message     string `json:"message"`
	From        string `json:"from"`
	CreatedTime string `json:"created_time"`
	LikeCount   int    `json:"like_count"`
	IsHidden    bool   `json:"is_hidden"`
}
