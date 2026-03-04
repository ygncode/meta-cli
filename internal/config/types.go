package config

type Config struct {
	DefaultAccount  string             `json:"default_account,omitempty"`
	DefaultPage     string             `json:"default_page,omitempty"`
	GraphAPIVersion string             `json:"graph_api_version,omitempty"`
	WebhookPort     int                `json:"webhook_port,omitempty"`
	RAGDir          string             `json:"rag_dir,omitempty"`
	DBPath          string             `json:"db_path,omitempty"`
	Accounts        map[string]Account `json:"accounts,omitempty"`
}

type Account struct {
	AppID string `json:"app_id"`
}

func Default() *Config {
	return &Config{
		DefaultAccount:  "default",
		GraphAPIVersion: "v21.0",
		WebhookPort:     8080,
	}
}
