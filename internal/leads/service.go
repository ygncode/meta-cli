package leads

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

func New(client *graph.Client) *Service {
	return &Service{client: client}
}

func (s *Service) CreateForm(ctx context.Context, pageID string, payload json.RawMessage) (string, error) {
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		return "", fmt.Errorf("invalid form JSON: %w", err)
	}

	body := url.Values{}
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			body.Set(k, val)
		default:
			data, _ := json.Marshal(v)
			body.Set(k, string(data))
		}
	}

	var result struct {
		ID string `json:"id"`
	}

	if err := s.client.Post(ctx, pageID+"/leadgen_forms", body, &result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (s *Service) ListLeads(ctx context.Context, formID string, limit int) ([]Lead, error) {
	var result struct {
		Data []struct {
			ID          string `json:"id"`
			CreatedTime string `json:"created_time"`
			FieldData   []struct {
				Name   string   `json:"name"`
				Values []string `json:"values"`
			} `json:"field_data"`
		} `json:"data"`
	}

	params := url.Values{
		"fields": {"id,created_time,field_data"},
		"limit":  {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, formID+"/leads", params, &result); err != nil {
		return nil, err
	}

	leads := make([]Lead, 0, len(result.Data))
	for _, d := range result.Data {
		fieldData, _ := json.Marshal(d.FieldData)
		leads = append(leads, Lead{
			ID:          d.ID,
			CreatedTime: d.CreatedTime,
			FieldData:   string(fieldData),
		})
	}
	return leads, nil
}
