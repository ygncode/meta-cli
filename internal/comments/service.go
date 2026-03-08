package comments

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ygncode/meta-cli/internal/graph"
)

type Service struct {
	client *graph.Client
}

func New(client *graph.Client) *Service {
	return &Service{client: client}
}

func (s *Service) List(ctx context.Context, postID string, limit int) ([]Comment, error) {
	var result struct {
		Data []struct {
			ID          string `json:"id"`
			Message     string `json:"message"`
			From        *struct {
				Name string `json:"name"`
			} `json:"from"`
			CreatedTime string `json:"created_time"`
			LikeCount   int    `json:"like_count"`
			IsHidden    bool   `json:"is_hidden"`
		} `json:"data"`
	}

	params := url.Values{
		"fields": {"id,message,from,created_time,like_count,is_hidden"},
		"limit":  {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, postID+"/comments", params, &result); err != nil {
		return nil, err
	}

	comments := make([]Comment, 0, len(result.Data))
	for _, d := range result.Data {
		c := Comment{
			ID:          d.ID,
			Message:     d.Message,
			CreatedTime: d.CreatedTime,
			LikeCount:   d.LikeCount,
			IsHidden:    d.IsHidden,
		}
		if d.From != nil {
			c.From = d.From.Name
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (s *Service) Reply(ctx context.Context, commentID, message string) (string, error) {
	var result struct {
		ID string `json:"id"`
	}
	body := url.Values{"message": {message}}
	if err := s.client.Post(ctx, commentID+"/comments", body, &result); err != nil {
		return "", err
	}
	return result.ID, nil
}

func (s *Service) SetHidden(ctx context.Context, commentID string, hidden bool) error {
	var result struct {
		Success bool `json:"success"`
	}
	body := url.Values{"is_hidden": {strconv.FormatBool(hidden)}}
	if err := s.client.Post(ctx, commentID, body, &result); err != nil {
		return err
	}
	if !result.Success {
		action := "hide"
		if !hidden {
			action = "unhide"
		}
		return fmt.Errorf("failed to %s comment %s", action, commentID)
	}
	return nil
}

func (s *Service) Update(ctx context.Context, commentID, message string) error {
	var result struct {
		Success bool `json:"success"`
	}
	body := url.Values{"message": {message}}
	if err := s.client.Post(ctx, commentID, body, &result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to update comment %s", commentID)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, commentID string) error {
	var result struct {
		Success bool `json:"success"`
	}
	if err := s.client.Delete(ctx, commentID, &result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to delete comment %s", commentID)
	}
	return nil
}
