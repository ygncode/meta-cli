package messenger_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ygncode/meta-cli/internal/messenger"
)

func makeSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookVerify(t *testing.T) {
	handler := &messenger.WebhookHandler{
		VerifyToken: "my_verify_token",
		AppSecret:   "secret",
	}

	t.Run("valid verification", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/webhook?hub.mode=subscribe&hub.verify_token=my_verify_token&hub.challenge=CHALLENGE",
			nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		if w.Body.String() != "CHALLENGE" {
			t.Errorf("expected CHALLENGE, got %s", w.Body.String())
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/webhook?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=CHALLENGE",
			nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("wrong mode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/webhook?hub.mode=unsubscribe&hub.verify_token=my_verify_token&hub.challenge=CHALLENGE",
			nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})
}

func TestWebhookInfoPage(t *testing.T) {
	handler := &messenger.WebhookHandler{
		VerifyToken: "my_verify_token",
		AppSecret:   "secret",
		PageID:      "page_123",
	}

	t.Run("plain GET returns status info", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if body["status"] != "running" {
			t.Errorf("expected status=running, got %v", body["status"])
		}
		if body["service"] != "meta-cli webhook" {
			t.Errorf("expected service=meta-cli webhook, got %v", body["service"])
		}
		if body["page_id"] != "page_123" {
			t.Errorf("expected page_id=page_123, got %v", body["page_id"])
		}
		if body["auto_reply"] != "disabled" {
			t.Errorf("expected auto_reply=disabled, got %v", body["auto_reply"])
		}
	})

	t.Run("plain GET with auto-reply enabled", func(t *testing.T) {
		deb := &mockDebouncer{}
		h := &messenger.WebhookHandler{
			VerifyToken: "tok",
			AppSecret:   "secret",
			PageID:      "page_456",
			Debouncer:   deb,
		}

		req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if body["auto_reply"] != "enabled" {
			t.Errorf("expected auto_reply=enabled, got %v", body["auto_reply"])
		}
		if body["page_id"] != "page_456" {
			t.Errorf("expected page_id=page_456, got %v", body["page_id"])
		}
	})

	t.Run("GET with random query params still shows info", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/webhook?foo=bar", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestWebhookReceiveValidSignature(t *testing.T) {
	store := openTestStore(t)

	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "test_secret",
		PageID:      "page_1",
		Store:       store,
	}

	payload := messenger.WebhookPayload{
		Object: "page",
		Entry: []messenger.Entry{
			{
				ID:   "page_1",
				Time: 1234567890000,
				Messaging: []messenger.Messaging{
					{
						Sender:    messenger.Participant{ID: "user_1"},
						Recipient: messenger.Participant{ID: "page_1"},
						Timestamp: 1234567890000,
						Message:   &messenger.MsgPayload{MID: "mid_100", Text: "Hello!"},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeSignature(body, "test_secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sig)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "EVENT_RECEIVED" {
		t.Errorf("expected EVENT_RECEIVED, got %s", w.Body.String())
	}
}

func TestWebhookReceiveInvalidSignature(t *testing.T) {
	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "test_secret",
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestWebhookReceiveMissingSignature(t *testing.T) {
	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "test_secret",
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestWebhookEmptyAppSecret(t *testing.T) {
	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "",
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	req.Header.Set("X-Hub-Signature-256", "sha256=anything")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for empty secret, got %d", w.Code)
	}
}

func TestWebhookMethodNotAllowed(t *testing.T) {
	handler := &messenger.WebhookHandler{}

	req := httptest.NewRequest(http.MethodPut, "/webhook", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

// mockDebouncer records calls to Add for testing.
type mockDebouncer struct {
	mu    sync.Mutex
	calls []debouncerCall
}

type debouncerCall struct {
	PSID string
	Msg  messenger.DebouncerMessage
}

func (m *mockDebouncer) Add(psid string, msg messenger.DebouncerMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, debouncerCall{PSID: psid, Msg: msg})
}

func (m *mockDebouncer) getCalls() []debouncerCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]debouncerCall(nil), m.calls...)
}

func TestWebhookWithDebouncer(t *testing.T) {
	store := openTestStore(t)
	deb := &mockDebouncer{}

	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "secret",
		PageID:      "page_1",
		Store:       store,
		Debouncer:   deb,
	}

	payload := messenger.WebhookPayload{
		Object: "page",
		Entry: []messenger.Entry{
			{
				ID:   "page_1",
				Time: 1234567890000,
				Messaging: []messenger.Messaging{
					{
						Sender:    messenger.Participant{ID: "user_1"},
						Recipient: messenger.Participant{ID: "page_1"},
						Timestamp: 1234567890000,
						Message:   &messenger.MsgPayload{MID: "mid_200", Text: "Hello debouncer!"},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeSignature(body, "secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sig)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// processPayload runs in a goroutine; wait for debouncer call
	var calls []debouncerCall
	for i := 0; i < 50; i++ {
		time.Sleep(10 * time.Millisecond)
		calls = deb.getCalls()
		if len(calls) > 0 {
			break
		}
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 debouncer call, got %d", len(calls))
	}
	if calls[0].PSID != "user_1" {
		t.Errorf("expected PSID user_1, got %s", calls[0].PSID)
	}
	if calls[0].Msg.Text != "Hello debouncer!" {
		t.Errorf("expected 'Hello debouncer!', got %s", calls[0].Msg.Text)
	}
}

func TestWebhookWithoutDebouncer(t *testing.T) {
	store := openTestStore(t)

	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "secret",
		PageID:      "page_1",
		Store:       store,
		// Debouncer is nil — backward compatible
	}

	payload := messenger.WebhookPayload{
		Object: "page",
		Entry: []messenger.Entry{
			{
				ID:   "page_1",
				Time: 1234567890000,
				Messaging: []messenger.Messaging{
					{
						Sender:    messenger.Participant{ID: "user_1"},
						Recipient: messenger.Participant{ID: "page_1"},
						Timestamp: 1234567890000,
						Message:   &messenger.MsgPayload{MID: "mid_300", Text: "No debouncer"},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeSignature(body, "secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sig)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Wait for message to be stored
	var msgs []messenger.Message
	for i := 0; i < 50; i++ {
		time.Sleep(10 * time.Millisecond)
		var err error
		msgs, err = store.ListMessages("page_1", 10)
		if err != nil {
			t.Fatalf("ListMessages: %v", err)
		}
		if len(msgs) > 0 {
			break
		}
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message stored, got %d", len(msgs))
	}
}

func TestWebhookEchoNotDebounced(t *testing.T) {
	store := openTestStore(t)
	deb := &mockDebouncer{}

	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "secret",
		PageID:      "page_1",
		Store:       store,
		Debouncer:   deb,
	}

	payload := messenger.WebhookPayload{
		Object: "page",
		Entry: []messenger.Entry{
			{
				ID:   "page_1",
				Time: 1234567890000,
				Messaging: []messenger.Messaging{
					{
						Sender:    messenger.Participant{ID: "page_1"},
						Recipient: messenger.Participant{ID: "user_1"},
						Timestamp: 1234567890000,
						Message:   &messenger.MsgPayload{MID: "mid_echo2", Text: "Echo msg", IsEcho: true},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeSignature(body, "secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sig)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Wait for processPayload goroutine to complete
	time.Sleep(100 * time.Millisecond)

	calls := deb.getCalls()
	if len(calls) != 0 {
		t.Errorf("expected 0 debouncer calls for echo messages, got %d", len(calls))
	}
}

func TestWebhookEchoMessage(t *testing.T) {
	store := openTestStore(t)

	handler := &messenger.WebhookHandler{
		VerifyToken: "tok",
		AppSecret:   "secret",
		PageID:      "page_1",
		Store:       store,
	}

	payload := messenger.WebhookPayload{
		Object: "page",
		Entry: []messenger.Entry{
			{
				ID:   "page_1",
				Time: 1234567890000,
				Messaging: []messenger.Messaging{
					{
						Sender:    messenger.Participant{ID: "page_1"},
						Recipient: messenger.Participant{ID: "user_1"},
						Timestamp: 1234567890000,
						Message:   &messenger.MsgPayload{MID: "mid_echo", Text: "Hi from page", IsEcho: true},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeSignature(body, "secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sig)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// processPayload runs in a goroutine; wait for it
	var msgs []messenger.Message
	for i := 0; i < 50; i++ {
		time.Sleep(10 * time.Millisecond)
		var err error
		msgs, err = store.ListMessages("page_1", 10)
		if err != nil {
			t.Fatalf("ListMessages: %v", err)
		}
		if len(msgs) > 0 {
			break
		}
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	msg := msgs[0]
	if msg.Direction != "out" {
		t.Errorf("expected direction out, got %s", msg.Direction)
	}
	if msg.PSID != "user_1" {
		t.Errorf("expected PSID user_1, got %s", msg.PSID)
	}
	if msg.ID != "mid_echo" {
		t.Errorf("expected ID mid_echo, got %s", msg.ID)
	}
}
