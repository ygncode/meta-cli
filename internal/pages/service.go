package pages

import (
	"context"
	"net/url"

	"github.com/ygncode/meta-cli/internal/graph"
)

type Page struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Service struct {
	client *graph.Client
}

func New(client *graph.Client) *Service {
	return &Service{client: client}
}

func (s *Service) List(ctx context.Context) ([]Page, error) {
	var result struct {
		Data []Page `json:"data"`
	}
	params := url.Values{"fields": {"id,name"}}
	if err := s.client.Get(ctx, "me/accounts", params, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}
