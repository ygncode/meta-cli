package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/auth"
)

func TestLoginURL(t *testing.T) {
	url := auth.LoginURL("12345", "v21.0", "")
	if !strings.Contains(url, "facebook.com/v21.0/dialog/oauth") {
		t.Errorf("expected OAuth dialog URL, got %s", url)
	}
	if !strings.Contains(url, "client_id=12345") {
		t.Errorf("expected client_id in URL, got %s", url)
	}
	if !strings.Contains(url, "response_type=code") {
		t.Errorf("expected response_type=code in URL, got %s", url)
	}
}

func TestExtractCode(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "code in query",
			url:  "https://localhost/?code=abc123",
			want: "abc123",
		},
		{
			name: "code in fragment",
			url:  "https://localhost/#code=xyz789",
			want: "xyz789",
		},
		{
			name:    "no code",
			url:     "https://localhost/?foo=bar",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "://bad",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auth.ExtractCode(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPageAccessToken(t *testing.T) {
	tokens := &auth.Tokens{
		Pages: map[string]auth.PageToken{
			"111": {Name: "Page A", Token: "tok_a"},
		},
	}

	tok, ok := tokens.PageAccessToken("111")
	if !ok || tok != "tok_a" {
		t.Errorf("expected tok_a, got %q ok=%v", tok, ok)
	}

	_, ok = tokens.PageAccessToken("999")
	if ok {
		t.Error("expected false for missing page")
	}

	// Nil tokens
	var nilTokens *auth.Tokens
	_, ok = nilTokens.PageAccessToken("111")
	if ok {
		t.Error("expected false for nil tokens")
	}
}

func TestPageNames(t *testing.T) {
	tokens := &auth.Tokens{
		Pages: map[string]auth.PageToken{
			"111": {Name: "Page A", Token: "tok_a"},
			"222": {Name: "Page B", Token: "tok_b"},
		},
	}

	names := tokens.PageNames()
	if names["111"] != "Page A" || names["222"] != "Page B" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestMemStore(t *testing.T) {
	store := auth.NewMemStore()

	// Initially empty
	_, err := store.GetTokens("default")
	if err == nil {
		t.Error("expected error for empty store")
	}

	_, err = store.GetSecret("default")
	if err == nil {
		t.Error("expected error for empty store")
	}

	// Save and retrieve tokens
	tokens := &auth.Tokens{
		UserToken: "user_tok",
		Pages: map[string]auth.PageToken{
			"111": {Name: "P", Token: "t"},
		},
	}
	if err := store.SaveTokens("default", tokens); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}
	got, err := store.GetTokens("default")
	if err != nil {
		t.Fatalf("GetTokens: %v", err)
	}
	if got.UserToken != "user_tok" {
		t.Errorf("expected user_tok, got %s", got.UserToken)
	}

	// Save and retrieve secret
	if err := store.SaveSecret("default", "my_secret"); err != nil {
		t.Fatalf("SaveSecret: %v", err)
	}
	secret, err := store.GetSecret("default")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if secret != "my_secret" {
		t.Errorf("expected my_secret, got %s", secret)
	}
}

func TestExchangeCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "test_code" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": "invalid code", "code": 100},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"access_token": "short_token"})
	}))
	defer srv.Close()

	// ExchangeCode uses hardcoded graph.facebook.com URL, so we can't easily redirect it.
	// Instead, test the error path with a cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := auth.ExchangeCode(ctx, "test_code", "app", "secret", "v21.0", "")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestExtendToken(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := auth.ExtendToken(ctx, "short", "app", "secret", "v21.0")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestFetchPageTokens(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := auth.FetchPageTokens(ctx, "token", "v21.0")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
