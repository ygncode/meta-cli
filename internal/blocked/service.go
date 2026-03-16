package blocked

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

func (s *Service) List(ctx context.Context, pageID string, limit int) ([]BlockedUser, error) {
	var result struct {
		Data []BlockedUser `json:"data"`
	}

	params := url.Values{
		"limit": {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, pageID+"/blocked", params, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *Service) Block(ctx context.Context, pageID, userID string) error {
	var result struct {
		Success bool `json:"success"`
	}

	body := url.Values{"user": {userID}}
	if err := s.client.Post(ctx, pageID+"/blocked", body, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to block user %s", userID)
	}
	return nil
}

func (s *Service) Unblock(ctx context.Context, pageID, userID string) error {
	var result struct {
		Success bool `json:"success"`
	}

	params := url.Values{"user": {userID}}
	if err := s.client.DeleteWithParams(ctx, pageID+"/blocked", params, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to unblock user %s", userID)
	}
	return nil
}
