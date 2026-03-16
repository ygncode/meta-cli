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

func (s *Service) Info(ctx context.Context, pageID string) (*PageInfo, error) {
	var raw struct {
		ID                 string   `json:"id"`
		Name               string   `json:"name"`
		About              string   `json:"about"`
		Description        string   `json:"description"`
		Category           string   `json:"category"`
		Phone              string   `json:"phone"`
		Website            string   `json:"website"`
		Emails             []string `json:"emails"`
		FanCount           int      `json:"fan_count"`
		FollowersCount     int      `json:"followers_count"`
		VerificationStatus string   `json:"verification_status"`
	}

	params := url.Values{
		"fields": {"id,name,about,description,category,phone,website,emails,fan_count,followers_count,verification_status"},
	}

	if err := s.client.Get(ctx, pageID, params, &raw); err != nil {
		return nil, err
	}

	return &PageInfo{
		ID:                 raw.ID,
		Name:               raw.Name,
		About:              raw.About,
		Description:        raw.Description,
		Category:           raw.Category,
		Phone:              raw.Phone,
		Website:            raw.Website,
		Emails:             raw.Emails,
		FanCount:           raw.FanCount,
		FollowersCount:     raw.FollowersCount,
		VerificationStatus: raw.VerificationStatus,
	}, nil
}
