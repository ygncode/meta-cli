package posts

import (
	"context"
	"fmt"
	"net/url"
	"time"

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

func validateScheduleTime(t time.Time) error {
	now := time.Now()
	if t.Before(now.Add(10 * time.Minute)) {
		return fmt.Errorf("schedule time must be at least 10 minutes in the future")
	}
	if t.After(now.Add(75 * 24 * time.Hour)) {
		return fmt.Errorf("schedule time must be within 75 days")
	}
	return nil
}

func applyScheduleOpts(body url.Values, opts *ScheduleOpts) error {
	if opts == nil {
		return nil
	}
	if err := validateScheduleTime(opts.PublishTime); err != nil {
		return err
	}
	body.Set("published", "false")
	body.Set("scheduled_publish_time", fmt.Sprintf("%d", opts.PublishTime.Unix()))
	return nil
}

func applyScheduleOptsToFields(fields map[string]string, opts *ScheduleOpts) error {
	if opts == nil {
		return nil
	}
	if err := validateScheduleTime(opts.PublishTime); err != nil {
		return err
	}
	fields["published"] = "false"
	fields["scheduled_publish_time"] = fmt.Sprintf("%d", opts.PublishTime.Unix())
	return nil
}

func (s *Service) CreateText(ctx context.Context, pageID, message string, opts *ScheduleOpts) (*CreateResult, error) {
	var result CreateResult
	body := url.Values{"message": {message}}
	if err := applyScheduleOpts(body, opts); err != nil {
		return nil, err
	}
	if err := s.client.Post(ctx, pageID+"/feed", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) CreatePhoto(ctx context.Context, pageID, message, photoPath string, opts *ScheduleOpts) (*CreateResult, error) {
	var result CreateResult
	fields := map[string]string{}
	if message != "" {
		fields["message"] = message
	}
	if err := applyScheduleOptsToFields(fields, opts); err != nil {
		return nil, err
	}
	if err := s.client.PostMultipart(ctx, pageID+"/photos", fields, photoPath, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) CreatePhotos(ctx context.Context, pageID, message string, photoPaths []string, opts *ScheduleOpts) (*CreateResult, error) {
	// Step 1: upload each photo as unpublished
	mediaIDs := make([]string, 0, len(photoPaths))
	for _, p := range photoPaths {
		var photoResult struct {
			ID string `json:"id"`
		}
		fields := map[string]string{"published": "false", "temporary": "true"}
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
	if err := applyScheduleOpts(body, opts); err != nil {
		return nil, err
	}

	var result CreateResult
	if err := s.client.Post(ctx, pageID+"/feed", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) CreateReel(ctx context.Context, pageID string, opts ReelOpts, schedOpts *ScheduleOpts) (*CreateResult, error) {
	// Step 1: Init upload session
	var initResult struct {
		VideoID   string `json:"video_id"`
		UploadURL string `json:"upload_url"`
	}
	if err := s.client.Post(ctx, pageID+"/video_reels", url.Values{"upload_phase": {"start"}}, &initResult); err != nil {
		return nil, fmt.Errorf("init reel upload: %w", err)
	}

	// Step 2: Upload binary
	var uploadResult struct {
		Success bool `json:"success"`
	}
	if err := s.client.PostBinary(ctx, initResult.UploadURL, opts.FilePath, &uploadResult); err != nil {
		return nil, fmt.Errorf("upload reel video: %w", err)
	}
	if !uploadResult.Success {
		return nil, fmt.Errorf("upload reel video: upload returned success=false")
	}

	// Step 3: Finish/publish
	finishBody := url.Values{
		"upload_phase": {"finish"},
		"video_id":     {initResult.VideoID},
	}
	if opts.Message != "" {
		finishBody.Set("description", opts.Message)
	}
	if opts.Title != "" {
		finishBody.Set("title", opts.Title)
	}
	if schedOpts != nil {
		if err := applyScheduleOpts(finishBody, schedOpts); err != nil {
			return nil, err
		}
	} else {
		finishBody.Set("published", "true")
	}

	var result CreateResult
	if err := s.client.Post(ctx, pageID+"/video_reels", finishBody, &result); err != nil {
		return nil, fmt.Errorf("finish reel publish: %w", err)
	}
	return &result, nil
}

func (s *Service) CreateVideo(ctx context.Context, pageID string, vopts VideoOpts, schedOpts *ScheduleOpts) (*CreateResult, error) {
	var result CreateResult
	fields := map[string]string{}
	if vopts.Message != "" {
		fields["description"] = vopts.Message
	}
	if vopts.Title != "" {
		fields["title"] = vopts.Title
	}
	if err := applyScheduleOptsToFields(fields, schedOpts); err != nil {
		return nil, err
	}
	files := map[string]string{"source": vopts.FilePath}
	if vopts.Thumbnail != "" {
		files["thumb"] = vopts.Thumbnail
	}
	if err := s.client.PostMultipartFiles(ctx, pageID+"/videos", fields, files, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) CreateLink(ctx context.Context, pageID, message, link string, opts *ScheduleOpts) (*CreateResult, error) {
	var result CreateResult
	body := url.Values{"link": {link}}
	if message != "" {
		body.Set("message", message)
	}
	if err := applyScheduleOpts(body, opts); err != nil {
		return nil, err
	}
	if err := s.client.Post(ctx, pageID+"/feed", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) Update(ctx context.Context, postID, message string) error {
	var result struct {
		Success bool `json:"success"`
	}
	body := url.Values{"message": {message}}
	if err := s.client.Post(ctx, postID, body, &result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to update post %s", postID)
	}
	return nil
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

func (s *Service) ListScheduled(ctx context.Context, pageID string, limit int) ([]ScheduledPost, error) {
	var result struct {
		Data []ScheduledPost `json:"data"`
	}

	params := url.Values{
		"fields": {"id,message,scheduled_publish_time,created_time"},
		"limit":  {fmt.Sprintf("%d", limit)},
	}

	if err := s.client.Get(ctx, pageID+"/scheduled_posts", params, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}
