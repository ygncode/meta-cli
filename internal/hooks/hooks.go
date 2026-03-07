package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/ygncode/meta-cli/internal/debounce"
)

// DefaultPromptTemplate is the Go template used when no custom template is configured.
const DefaultPromptTemplate = `New message(s) from Facebook Messenger user (PSID: {{.PSID}}) on page {{.PageID}}:

{{range .Messages}}- {{.Text}}
{{end}}
Use the meta-cli skill to help this user. Search the knowledge base if needed, check conversation history for context, and send a helpful reply.`

// Client calls the OpenClaw /hooks/agent endpoint.
type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

// agentRequest is the JSON body sent to /hooks/agent.
type agentRequest struct {
	Message    string `json:"message"`
	Name       string `json:"name"`
	Deliver    bool   `json:"deliver"`
	SessionKey string `json:"sessionKey"`
}

// agentResponse is the JSON response from /hooks/agent.
type agentResponse struct {
	OK    bool   `json:"ok"`
	RunID string `json:"runId"`
}

// NewClient creates a hooks client with the default HTTP client.
func NewClient(endpoint, token string) *Client {
	return &Client{
		endpoint:   endpoint,
		token:      token,
		httpClient: http.DefaultClient,
	}
}

// NewClientWithHTTP creates a hooks client with a custom HTTP client (for testing).
func NewClientWithHTTP(endpoint, token string, hc *http.Client) *Client {
	return &Client{
		endpoint:   endpoint,
		token:      token,
		httpClient: hc,
	}
}

// CallAgent sends a prompt to the OpenClaw /hooks/agent endpoint.
// It fires asynchronously on the server side; this call returns as soon as
// the server acknowledges the request.
func (c *Client) CallAgent(ctx context.Context, prompt string, psid string) error {
	reqBody := agentRequest{
		Message:    prompt,
		Name:       "FB Messenger",
		Deliver:    false,
		SessionKey: "hook:fb:" + psid,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal hooks request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create hooks request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hooks request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hooks/agent returned status %d", resp.StatusCode)
	}

	var result agentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode hooks response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("hooks/agent returned ok=false")
	}

	return nil
}

// PromptData is the data passed to the prompt template.
type PromptData struct {
	PSID     string
	PageID   string
	Messages []debounce.Message
}

// RenderPrompt renders a Go text template with the given data.
// If tmpl is empty, DefaultPromptTemplate is used.
func RenderPrompt(tmpl string, psid, pageID string, messages []debounce.Message) (string, error) {
	if tmpl == "" {
		tmpl = DefaultPromptTemplate
	}

	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}

	data := PromptData{
		PSID:     psid,
		PageID:   pageID,
		Messages: messages,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute prompt template: %w", err)
	}

	return buf.String(), nil
}
