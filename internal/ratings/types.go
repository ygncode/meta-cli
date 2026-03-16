package ratings

type Rating struct {
	ReviewerName string `json:"reviewer_name"`
	Rating       int    `json:"rating"`
	ReviewText   string `json:"review_text"`
	CreatedTime  string `json:"created_time"`
}

type OverallRating struct {
	StarRating  float64 `json:"star_rating"`
	RatingCount int     `json:"rating_count"`
}
