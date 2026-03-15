package insights

import (
	"context"
	"net/url"

	"github.com/ygncode/meta-cli/internal/graph"
)

type Service struct {
	client *graph.Client
}

func New(client *graph.Client) *Service {
	return &Service{client: client}
}

// GetPageInsights fetches page-level insights for the given metrics and period.
func (s *Service) GetPageInsights(ctx context.Context, pageID, metrics, period string) ([]InsightMetric, error) {
	var result struct {
		Data []InsightMetric `json:"data"`
	}

	params := url.Values{
		"metric": {metrics},
		"period": {period},
	}

	if err := s.client.Get(ctx, pageID+"/insights", params, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetPostInsights fetches post-level insights for the given metrics.
func (s *Service) GetPostInsights(ctx context.Context, postID, metrics string) ([]InsightMetric, error) {
	var result struct {
		Data []InsightMetric `json:"data"`
	}

	params := url.Values{
		"metric": {metrics},
	}

	if err := s.client.Get(ctx, postID+"/insights", params, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// Flatten converts a slice of InsightMetric into flat InsightRow entries for tabular output.
func Flatten(metrics []InsightMetric) []InsightRow {
	var rows []InsightRow
	for _, m := range metrics {
		for _, v := range m.Values {
			rows = append(rows, InsightRow{
				Metric:  m.Name,
				Period:  m.Period,
				EndTime: v.EndTime,
				Value:   v.Value,
			})
		}
	}
	return rows
}
