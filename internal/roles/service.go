package roles

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

func New(client *graph.Client) *Service {
	return &Service{client: client}
}

func (s *Service) List(ctx context.Context, pageID string) ([]AssignedUser, error) {
	var result struct {
		Data []struct {
			ID    string   `json:"id"`
			Name  string   `json:"name"`
			Tasks []string `json:"tasks"`
		} `json:"data"`
	}

	params := url.Values{
		"fields": {"id,name,tasks"},
	}

	if err := s.client.Get(ctx, pageID+"/assigned_users", params, &result); err != nil {
		return nil, err
	}

	users := make([]AssignedUser, 0, len(result.Data))
	for _, d := range result.Data {
		users = append(users, AssignedUser{
			ID:    d.ID,
			Name:  d.Name,
			Tasks: strings.Join(d.Tasks, ","),
		})
	}
	return users, nil
}

func (s *Service) Assign(ctx context.Context, pageID, userID string, tasks []string) error {
	var result struct {
		Success bool `json:"success"`
	}

	tasksJSON, _ := json.Marshal(tasks)
	body := url.Values{
		"user":  {userID},
		"tasks": {string(tasksJSON)},
	}

	if err := s.client.Post(ctx, pageID+"/assigned_users", body, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to assign user %s", userID)
	}
	return nil
}

func (s *Service) Remove(ctx context.Context, pageID, userID string) error {
	var result struct {
		Success bool `json:"success"`
	}

	params := url.Values{"user": {userID}}
	if err := s.client.DeleteWithParams(ctx, pageID+"/assigned_users", params, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to remove user %s", userID)
	}
	return nil
}
