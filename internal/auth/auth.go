package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var Scopes = []string{
	"pages_show_list",
	"pages_read_engagement",
	"pages_read_user_content",
	"pages_manage_posts",
	"pages_manage_engagement",
	"pages_messaging",
	"pages_manage_metadata",
	"public_profile",
}

func LoginURL(appID, version string) string {
	params := url.Values{
		"client_id":     {appID},
		"redirect_uri":  {"https://localhost/"},
		"scope":         {strings.Join(Scopes, ",")},
		"response_type": {"code"},
	}
	return fmt.Sprintf("https://www.facebook.com/%s/dialog/oauth?%s", version, params.Encode())
}

func ExtractCode(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	code := u.Query().Get("code")
	if code == "" {
		frag, _ := url.ParseQuery(u.Fragment)
		code = frag.Get("code")
	}
	if code == "" {
		return "", fmt.Errorf("no code parameter found in URL")
	}
	return code, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type graphErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func fetchToken(ctx context.Context, params url.Values, version string) (string, error) {
	u := fmt.Sprintf("https://graph.facebook.com/%s/oauth/access_token?%s", version, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp graphErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("api error %d: %s", errResp.Error.Code, errResp.Error.Message)
		}
		return "", fmt.Errorf("http error %d", resp.StatusCode)
	}

	var result tokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}
	return result.AccessToken, nil
}

func ExchangeCode(ctx context.Context, code, appID, appSecret, version string) (string, error) {
	params := url.Values{
		"client_id":     {appID},
		"client_secret": {appSecret},
		"redirect_uri":  {"https://localhost/"},
		"code":          {code},
	}
	return fetchToken(ctx, params, version)
}

func ExtendToken(ctx context.Context, shortToken, appID, appSecret, version string) (string, error) {
	params := url.Values{
		"grant_type":        {"fb_exchange_token"},
		"client_id":         {appID},
		"client_secret":     {appSecret},
		"fb_exchange_token": {shortToken},
	}
	return fetchToken(ctx, params, version)
}

type pageResponse struct {
	Data []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

func FetchPageTokens(ctx context.Context, userToken, version string) (map[string]PageToken, error) {
	params := url.Values{
		"access_token": {userToken},
		"fields":       {"id,name,access_token"},
	}

	u := fmt.Sprintf("https://graph.facebook.com/%s/me/accounts?%s", version, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp graphErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("api error %d: %s", errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("http error %d", resp.StatusCode)
	}

	var result pageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode pages response: %w", err)
	}

	pages := make(map[string]PageToken)
	for _, p := range result.Data {
		pages[p.ID] = PageToken{Name: p.Name, Token: p.AccessToken}
	}
	return pages, nil
}
