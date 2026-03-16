package pages

type PageInfo struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	About              string   `json:"about"`
	Description        string   `json:"description"`
	Category           string   `json:"category"`
	Phone              string   `json:"phone"`
	Website            string   `json:"website"`
	Emails             []string `json:"emails"`
	FanCount           int      `json:"fan_count"`
	FollowersCount     int      `json:"followers_count"`
	VerificationStatus string   `json:"verification_status"`
}
