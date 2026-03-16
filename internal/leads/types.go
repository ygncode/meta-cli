package leads

type Lead struct {
	ID          string `json:"id"`
	CreatedTime string `json:"created_time"`
	FieldData   string `json:"field_data"`
}

type Form struct {
	ID string `json:"id"`
}
