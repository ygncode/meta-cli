package ratings

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

func (s *Service) List(ctx context.Context, pageID string, limit int) ([]Rating, error) {
	var result struct {
		Data []struct {
			Reviewer *struct {
				Name string `json:"name"`
			} `json:"reviewer"`
			Rating      int    `json:"rating"`
			ReviewText  string `json:"review_text"`
			CreatedTime string `json:"created_time"`
		} `json:"data"`
	}

	params := url.Values{
		"fields": {"reviewer,rating,review_text,created_time"},
		"limit":  {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, pageID+"/ratings", params, &result); err != nil {
		return nil, err
	}

	ratings := make([]Rating, 0, len(result.Data))
	for _, d := range result.Data {
		r := Rating{
			Rating:      d.Rating,
			ReviewText:  d.ReviewText,
			CreatedTime: d.CreatedTime,
		}
		if d.Reviewer != nil {
			r.ReviewerName = d.Reviewer.Name
		}
		ratings = append(ratings, r)
	}
	return ratings, nil
}

func (s *Service) Summary(ctx context.Context, pageID string) (*OverallRating, error) {
	var result struct {
		OverallStarRating float64 `json:"overall_star_rating"`
		RatingCount       int     `json:"rating_count"`
	}

	params := url.Values{
		"fields": {"overall_star_rating,rating_count"},
	}

	if err := s.client.Get(ctx, pageID, params, &result); err != nil {
		return nil, err
	}

	return &OverallRating{
		StarRating:  result.OverallStarRating,
		RatingCount: result.RatingCount,
	}, nil
}
