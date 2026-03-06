package messenger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ygncode/meta-cli/internal/graph"
)

type Service struct {
	client *graph.Client
}

func NewService(client *graph.Client) *Service {
	return &Service{client: client}
}

func (s *Service) Send(ctx context.Context, psid, text string) (string, error) {
	body := url.Values{
		"messaging_type": {"RESPONSE"},
		"recipient":      {mustJSON(map[string]string{"id": psid})},
		"message":        {mustJSON(map[string]string{"text": text})},
	}

	var result struct {
		RecipientID string `json:"recipient_id"`
		MessageID   string `json:"message_id"`
	}
	if err := s.client.Post(ctx, "me/messages", body, &result); err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}
	return result.MessageID, nil
}

func (s *Service) SendTyping(ctx context.Context, psid string, on bool) error {
	action := "typing_on"
	if !on {
		action = "typing_off"
	}
	body := url.Values{
		"recipient":     {mustJSON(map[string]string{"id": psid})},
		"sender_action": {action},
	}

	if err := s.client.Post(ctx, "me/messages", body, &struct{}{}); err != nil {
		return fmt.Errorf("send typing: %w", err)
	}
	return nil
}

func (s *Service) SubscribeWebhook(ctx context.Context) error {
	body := url.Values{
		"subscribed_fields": {"messages,message_echoes"},
	}
	var result struct {
		Success bool `json:"success"`
	}
	if err := s.client.Post(ctx, "me/subscribed_apps", body, &result); err != nil {
		return fmt.Errorf("subscribe webhook: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("subscribe webhook: API returned success=false")
	}
	return nil
}

func mustJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
