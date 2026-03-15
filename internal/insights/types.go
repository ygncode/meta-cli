package insights

type InsightValue struct {
	Value   int    `json:"value"`
	EndTime string `json:"end_time"`
}

type InsightMetric struct {
	Name        string         `json:"name"`
	Period      string         `json:"period"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ID          string         `json:"id"`
	Values      []InsightValue `json:"values"`
}

type InsightRow struct {
	Metric  string `json:"metric"`
	Period  string `json:"period"`
	EndTime string `json:"end_time"`
	Value   int    `json:"value"`
}
