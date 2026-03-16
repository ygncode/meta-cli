package events

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

func (s *Service) List(ctx context.Context, pageID string, limit int) ([]Event, error) {
	var result struct {
		Data []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			StartTime   string `json:"start_time"`
			EndTime     string `json:"end_time"`
			Place       *struct {
				Name string `json:"name"`
			} `json:"place"`
		} `json:"data"`
	}

	params := url.Values{
		"fields": {"id,name,description,start_time,end_time,place"},
		"limit":  {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, pageID+"/events", params, &result); err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(result.Data))
	for _, d := range result.Data {
		e := Event{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
			StartTime:   d.StartTime,
			EndTime:     d.EndTime,
		}
		if d.Place != nil {
			e.Place = d.Place.Name
		}
		events = append(events, e)
	}
	return events, nil
}
