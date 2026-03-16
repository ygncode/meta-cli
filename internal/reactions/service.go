package reactions

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

func (s *Service) List(ctx context.Context, objectID string, limit int) ([]Reaction, error) {
	var result struct {
		Data []Reaction `json:"data"`
	}

	params := url.Values{
		"fields": {"id,name,type"},
		"limit":  {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, objectID+"/reactions", params, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
