package config

type Config struct {
	DefaultAccount  string             `json:"default_account,omitempty"`
	DefaultPage     string             `json:"default_page,omitempty"`
	GraphAPIVersion string             `json:"graph_api_version,omitempty"`
	WebhookPort     int                `json:"webhook_port,omitempty"`
	RAGDir          string             `json:"rag_dir,omitempty"`
	DBPath          string             `json:"db_path,omitempty"`
	VerifyToken     string             `json:"verify_token,omitempty"`
	Accounts        map[string]Account `json:"accounts,omitempty"`

	// Auto-reply pipeline fields
	DebounceSeconds int    `json:"debounce_seconds,omitempty"`
	HooksEndpoint   string `json:"hooks_endpoint,omitempty"`
	HooksToken      string `json:"hooks_token,omitempty"`
	AutoReply       bool   `json:"auto_reply,omitempty"`
	PromptTemplate  string `json:"prompt_template,omitempty"`
}

type Account struct {
	AppID string `json:"app_id"`
}

func Default() *Config {
	return &Config{
		DefaultAccount:  "default",
		GraphAPIVersion: "v25.0",
		WebhookPort:     8080,
		DebounceSeconds: 3,
	}
}
