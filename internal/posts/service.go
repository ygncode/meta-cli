package posts

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

func (s *Service) List(ctx context.Context, pageID string, limit int) ([]Post, error) {
	var result struct {
		Data []struct {
			ID           string `json:"id"`
			Message      string `json:"message"`
			CreatedTime  string `json:"created_time"`
			PermalinkURL string `json:"permalink_url"`
			Likes        *struct {
				Summary struct {
					TotalCount int `json:"total_count"`
				} `json:"summary"`
			} `json:"likes"`
			Comments *struct {
				Summary struct {
					TotalCount int `json:"total_count"`
				} `json:"summary"`
			} `json:"comments"`
			Shares *struct {
				Count int `json:"count"`
			} `json:"shares"`
		} `json:"data"`
	}

	params := url.Values{
		"fields": {"id,message,created_time,permalink_url,shares,likes.summary(true),comments.summary(true)"},
		"limit":  {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, pageID+"/posts", params, &result); err != nil {
		return nil, err
	}

	posts := make([]Post, 0, len(result.Data))
	for _, d := range result.Data {
		p := Post{
			ID:           d.ID,
			Message:      d.Message,
			CreatedTime:  d.CreatedTime,
			PermalinkURL: d.PermalinkURL,
		}
		if d.Likes != nil {
			p.Likes = d.Likes.Summary.TotalCount
		}
		if d.Comments != nil {
			p.Comments = d.Comments.Summary.TotalCount
		}
		if d.Shares != nil {
			p.Shares = d.Shares.Count
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (s *Service) CreateText(ctx context.Context, pageID, message string) (*CreateResult, error) {
	var result CreateResult
	body := url.Values{"message": {message}}
	if err := s.client.Post(ctx, pageID+"/feed", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) CreatePhoto(ctx context.Context, pageID, message, photoPath string) (*CreateResult, error) {
	var result CreateResult
	fields := map[string]string{}
	if message != "" {
		fields["message"] = message
	}
	if err := s.client.PostMultipart(ctx, pageID+"/photos", fields, photoPath, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) CreatePhotos(ctx context.Context, pageID, message string, photoPaths []string) (*CreateResult, error) {
	// Step 1: upload each photo as unpublished
	mediaIDs := make([]string, 0, len(photoPaths))
	for _, p := range photoPaths {
		var photoResult struct {
			ID string `json:"id"`
		}
		fields := map[string]string{"published": "false"}
		if err := s.client.PostMultipart(ctx, pageID+"/photos", fields, p, &photoResult); err != nil {
			return nil, fmt.Errorf("upload %s: %w", p, err)
		}
		mediaIDs = append(mediaIDs, photoResult.ID)
	}

	// Step 2: create feed post with attached_media
	body := url.Values{}
	if message != "" {
		body.Set("message", message)
	}
	for i, id := range mediaIDs {
		body.Set(fmt.Sprintf("attached_media[%d]", i), fmt.Sprintf("{\"media_fbid\":\"%s\"}", id))
	}

	var result CreateResult
	if err := s.client.Post(ctx, pageID+"/feed", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) CreateLink(ctx context.Context, pageID, message, link string) (*CreateResult, error) {
	var result CreateResult
	body := url.Values{"link": {link}}
	if message != "" {
		body.Set("message", message)
	}
	if err := s.client.Post(ctx, pageID+"/feed", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) Delete(ctx context.Context, postID string) error {
	var result struct {
		Success bool `json:"success"`
	}
	if err := s.client.Delete(ctx, postID, &result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to delete post %s", postID)
	}
	return nil
}
