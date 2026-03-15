package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

func New(version, token string) *Client {
	return &Client{
		baseURL:    "https://graph.facebook.com/" + version,
		httpClient: http.DefaultClient,
		token:      token,
	}
}

func NewWithHTTPClient(baseURL, token string, hc *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: hc,
		token:      token,
	}
}

func (c *Client) WithToken(token string) *Client {
	return &Client{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		token:      token,
	}
}

func (c *Client) Get(ctx context.Context, path string, params url.Values, out any) error {
	u, err := url.Parse(c.baseURL + "/" + path)
	if err != nil {
		return err
	}
	if params == nil {
		params = url.Values{}
	}
	params.Set("access_token", c.token)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) Post(ctx context.Context, path string, body url.Values, out any) error {
	u := c.baseURL + "/" + path + "?access_token=" + url.QueryEscape(c.token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, out)
}

func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string, filePath string, out any) error {
	return c.PostMultipartFiles(ctx, path, fields, map[string]string{"source": filePath}, out)
}

func (c *Client) PostMultipartFiles(ctx context.Context, path string, fields map[string]string, files map[string]string, out any) error {
	u := c.baseURL + "/" + path + "?access_token=" + url.QueryEscape(c.token)

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		for k, v := range fields {
			if err := writer.WriteField(k, v); err != nil {
				pw.CloseWithError(err)
				return
			}
		}

		for fieldName, filePath := range files {
			f, err := os.Open(filePath)
			if err != nil {
				pw.CloseWithError(err)
				return
			}

			part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
			if err != nil {
				f.Close()
				pw.CloseWithError(err)
				return
			}
			if _, err := io.Copy(part, f); err != nil {
				f.Close()
				pw.CloseWithError(err)
				return
			}
			f.Close()
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, pr)
	if err != nil {
		pr.CloseWithError(err)
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.do(req, out)
}

func (c *Client) PostBinary(ctx context.Context, uploadURL, filePath string, out any) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, f)
	if err != nil {
		return err
	}
	req.ContentLength = fi.Size()
	req.Header.Set("Authorization", "OAuth "+c.token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("offset", "0")
	req.Header.Set("file_size", fmt.Sprintf("%d", fi.Size()))
	return c.do(req, out)
}

func (c *Client) Delete(ctx context.Context, path string, out any) error {
	u := c.baseURL + "/" + path + "?access_token=" + url.QueryEscape(c.token)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) DeleteWithParams(ctx context.Context, path string, params url.Values, out any) error {
	u, err := url.Parse(c.baseURL + "/" + path)
	if err != nil {
		return err
	}
	if params == nil {
		params = url.Values{}
	}
	params.Set("access_token", c.token)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var envelope struct {
			Error *GraphError `json:"error"`
		}
		if json.Unmarshal(body, &envelope) == nil && envelope.Error != nil {
			return &APIError{StatusCode: resp.StatusCode, Graph: envelope.Error}
		}
		return &APIError{StatusCode: resp.StatusCode}
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
