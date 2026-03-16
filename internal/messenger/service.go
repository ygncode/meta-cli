package messenger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

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

// SendAttachmentURL sends a media attachment via URL.
func (s *Service) SendAttachmentURL(ctx context.Context, psid, attachType, mediaURL string) (string, error) {
	body := url.Values{
		"messaging_type": {"RESPONSE"},
		"recipient":      {mustJSON(map[string]string{"id": psid})},
		"message": {mustJSON(map[string]any{
			"attachment": map[string]any{
				"type":    attachType,
				"payload": map[string]any{"url": mediaURL},
			},
		})},
	}

	var result struct {
		RecipientID string `json:"recipient_id"`
		MessageID   string `json:"message_id"`
	}
	if err := s.client.Post(ctx, "me/messages", body, &result); err != nil {
		return "", fmt.Errorf("send attachment URL: %w", err)
	}
	return result.MessageID, nil
}

// SendAttachmentFile sends a local file as an attachment via multipart upload.
func (s *Service) SendAttachmentFile(ctx context.Context, psid, attachType, filePath string) (string, error) {
	var result struct {
		RecipientID string `json:"recipient_id"`
		MessageID   string `json:"message_id"`
	}

	fields := map[string]string{
		"messaging_type": "RESPONSE",
		"recipient":      mustJSON(map[string]string{"id": psid}),
		"message": mustJSON(map[string]any{
			"attachment": map[string]any{
				"type":    attachType,
				"payload": map[string]any{},
			},
		}),
	}
	files := map[string]string{"filedata": filePath}

	if err := s.client.PostMultipartFiles(ctx, "me/messages", fields, files, &result); err != nil {
		return "", fmt.Errorf("send attachment file: %w", err)
	}
	return result.MessageID, nil
}

// SendTagged sends a message using a message tag (for outside 24-hour window).
func (s *Service) SendTagged(ctx context.Context, psid, text, tag string) (string, error) {
	validTags := map[string]bool{
		"HUMAN_AGENT":            true,
		"ACCOUNT_UPDATE":         true,
		"POST_PURCHASE_UPDATE":   true,
		"CONFIRMED_EVENT_UPDATE": true,
	}
	if !validTags[tag] {
		return "", fmt.Errorf("invalid tag %q, must be one of: HUMAN_AGENT, ACCOUNT_UPDATE, POST_PURCHASE_UPDATE, CONFIRMED_EVENT_UPDATE", tag)
	}

	body := url.Values{
		"messaging_type": {"MESSAGE_TAG"},
		"tag":            {tag},
		"recipient":      {mustJSON(map[string]string{"id": psid})},
		"message":        {mustJSON(map[string]string{"text": text})},
	}

	var result struct {
		RecipientID string `json:"recipient_id"`
		MessageID   string `json:"message_id"`
	}
	if err := s.client.Post(ctx, "me/messages", body, &result); err != nil {
		return "", fmt.Errorf("send tagged message: %w", err)
	}
	return result.MessageID, nil
}

// SendPrivateReply sends a private Messenger reply to a comment.
func (s *Service) SendPrivateReply(ctx context.Context, commentID, text string) (string, error) {
	body := url.Values{
		"messaging_type": {"RESPONSE"},
		"recipient":      {mustJSON(map[string]string{"comment_id": commentID})},
		"message":        {mustJSON(map[string]string{"text": text})},
	}

	var result struct {
		RecipientID string `json:"recipient_id"`
		MessageID   string `json:"message_id"`
	}
	if err := s.client.Post(ctx, "me/messages", body, &result); err != nil {
		return "", fmt.Errorf("send private reply: %w", err)
	}
	return result.MessageID, nil
}

// SendTemplate sends a structured template message.
func (s *Service) SendTemplate(ctx context.Context, psid string, payload json.RawMessage) (string, error) {
	body := url.Values{
		"messaging_type": {"RESPONSE"},
		"recipient":      {mustJSON(map[string]string{"id": psid})},
		"message": {mustJSON(map[string]any{
			"attachment": map[string]any{
				"type":    "template",
				"payload": json.RawMessage(payload),
			},
		})},
	}

	var result struct {
		RecipientID string `json:"recipient_id"`
		MessageID   string `json:"message_id"`
	}
	if err := s.client.Post(ctx, "me/messages", body, &result); err != nil {
		return "", fmt.Errorf("send template: %w", err)
	}
	return result.MessageID, nil
}

// SendWithQuickReplies sends a text message with quick reply buttons.
func (s *Service) SendWithQuickReplies(ctx context.Context, psid, text string, replies []string) (string, error) {
	qr := make([]map[string]string, 0, len(replies))
	for _, r := range replies {
		qr = append(qr, map[string]string{
			"content_type": "text",
			"title":        r,
			"payload":      r,
		})
	}

	body := url.Values{
		"messaging_type": {"RESPONSE"},
		"recipient":      {mustJSON(map[string]string{"id": psid})},
		"message": {mustJSON(map[string]any{
			"text":          text,
			"quick_replies": qr,
		})},
	}

	var result struct {
		RecipientID string `json:"recipient_id"`
		MessageID   string `json:"message_id"`
	}
	if err := s.client.Post(ctx, "me/messages", body, &result); err != nil {
		return "", fmt.Errorf("send quick replies: %w", err)
	}
	return result.MessageID, nil
}

// GetProfile returns the current Messenger profile configuration.
func (s *Service) GetProfile(ctx context.Context) (json.RawMessage, error) {
	params := url.Values{
		"fields": {"greeting,get_started,persistent_menu,ice_breakers"},
	}

	var result json.RawMessage
	if err := s.client.Get(ctx, "me/messenger_profile", params, &result); err != nil {
		return nil, fmt.Errorf("get messenger profile: %w", err)
	}
	return result, nil
}

// SetGreeting sets the Messenger greeting text.
func (s *Service) SetGreeting(ctx context.Context, text string) error {
	body := url.Values{
		"greeting": {mustJSON([]map[string]string{
			{"locale": "default", "text": text},
		})},
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := s.client.Post(ctx, "me/messenger_profile", body, &result); err != nil {
		return fmt.Errorf("set greeting: %w", err)
	}
	return nil
}

// SetGetStarted sets the Get Started button payload.
func (s *Service) SetGetStarted(ctx context.Context, payload string) error {
	body := url.Values{
		"get_started": {mustJSON(map[string]string{"payload": payload})},
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := s.client.Post(ctx, "me/messenger_profile", body, &result); err != nil {
		return fmt.Errorf("set get started: %w", err)
	}
	return nil
}

// SetMenu sets the persistent menu configuration.
func (s *Service) SetMenu(ctx context.Context, menu json.RawMessage) error {
	body := url.Values{
		"persistent_menu": {string(menu)},
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := s.client.Post(ctx, "me/messenger_profile", body, &result); err != nil {
		return fmt.Errorf("set menu: %w", err)
	}
	return nil
}

// SetIceBreakers sets the ice breaker conversation starters.
func (s *Service) SetIceBreakers(ctx context.Context, iceBreakers json.RawMessage) error {
	body := url.Values{
		"ice_breakers": {string(iceBreakers)},
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := s.client.Post(ctx, "me/messenger_profile", body, &result); err != nil {
		return fmt.Errorf("set ice breakers: %w", err)
	}
	return nil
}

// DeleteProfileField removes a field from the Messenger profile.
func (s *Service) DeleteProfileField(ctx context.Context, field string) error {
	body := url.Values{
		"fields": {mustJSON([]string{field})},
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := s.client.Post(ctx, "me/messenger_profile", body, &result); err != nil {
		return fmt.Errorf("delete profile field: %w", err)
	}
	return nil
}

// ListConversations lists Messenger conversations from the API.
func (s *Service) ListConversations(ctx context.Context, pageID string, limit int) ([]Conversation, error) {
	var result struct {
		Data []struct {
			ID           string `json:"id"`
			UpdatedTime  string `json:"updated_time"`
			MessageCount int    `json:"message_count"`
			Participants struct {
				Data []struct {
					Name string `json:"name"`
				} `json:"data"`
			} `json:"participants"`
		} `json:"data"`
	}

	params := url.Values{
		"platform": {"MESSENGER"},
		"fields":   {"id,participants,updated_time,message_count"},
		"limit":    {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, pageID+"/conversations", params, &result); err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}

	convos := make([]Conversation, 0, len(result.Data))
	for _, d := range result.Data {
		names := make([]string, 0, len(d.Participants.Data))
		for _, p := range d.Participants.Data {
			names = append(names, p.Name)
		}
		convos = append(convos, Conversation{
			ID:           d.ID,
			Participants: strings.Join(names, ", "),
			UpdatedTime:  d.UpdatedTime,
			MessageCount: d.MessageCount,
		})
	}
	return convos, nil
}

func mustJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
