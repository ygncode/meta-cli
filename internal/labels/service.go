package labels

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ygncode/meta-cli/internal/graph"
)

type Service struct {
	client *graph.Client
}

func New(client *graph.Client) *Service {
	return &Service{client: client}
}

// List returns all custom labels for a page.
func (s *Service) List(ctx context.Context, pageID string) ([]Label, error) {
	var result struct {
		Data []Label `json:"data"`
	}

	params := url.Values{
		"fields": {"id,name"},
	}

	if err := s.client.Get(ctx, pageID+"/custom_labels", params, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// Create creates a new custom label on a page.
func (s *Service) Create(ctx context.Context, pageID, name string) (string, error) {
	var result struct {
		ID string `json:"id"`
	}

	body := url.Values{"name": {name}}
	if err := s.client.Post(ctx, pageID+"/custom_labels", body, &result); err != nil {
		return "", err
	}

	return result.ID, nil
}

// Delete removes a custom label.
func (s *Service) Delete(ctx context.Context, labelID string) error {
	var result struct {
		Success bool `json:"success"`
	}

	if err := s.client.Delete(ctx, labelID, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to delete label %s", labelID)
	}
	return nil
}

// Assign assigns a label to a user (by PSID).
func (s *Service) Assign(ctx context.Context, labelID, userPSID string) error {
	var result struct {
		Success bool `json:"success"`
	}

	body := url.Values{"user": {userPSID}}
	if err := s.client.Post(ctx, labelID+"/label", body, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to assign label %s to user %s", labelID, userPSID)
	}
	return nil
}

// Remove removes a label from a user (by PSID).
func (s *Service) Remove(ctx context.Context, labelID, userPSID string) error {
	var result struct {
		Success bool `json:"success"`
	}

	params := url.Values{"user": {userPSID}}
	// The Graph API uses DELETE with query params for label removal
	if err := s.client.DeleteWithParams(ctx, labelID+"/label", params, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to remove label %s from user %s", labelID, userPSID)
	}
	return nil
}

// ListByUser returns all labels assigned to a specific user (by PSID).
func (s *Service) ListByUser(ctx context.Context, userPSID string) ([]Label, error) {
	var result struct {
		Data []Label `json:"data"`
	}

	params := url.Values{
		"fields": {"id,name"},
	}

	if err := s.client.Get(ctx, userPSID+"/custom_labels", params, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
