package messenger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// DebouncerInterface abstracts the debounce.Debouncer for testability.
type DebouncerInterface interface {
	Add(psid string, msg DebouncerMessage)
}

// DebouncerMessage matches debounce.Message fields.
type DebouncerMessage struct {
	ID   string
	Text string
}

type WebhookHandler struct {
	VerifyToken string
	AppSecret   string
	PageID      string
	Store       *Store
	Messenger   *Service
	Debouncer   DebouncerInterface // nil = auto-reply disabled
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.verify(w, r)
	case http.MethodPost:
		h.receive(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *WebhookHandler) verify(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == h.VerifyToken {
		log.Printf("Webhook verified")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, challenge)
		return
	}
	log.Printf("Webhook verification failed: mode=%s token_match=%v", mode, token == h.VerifyToken)
	w.WriteHeader(http.StatusForbidden)
}

func (h *WebhookHandler) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if !h.validateSignature(body, sig) {
		log.Printf("Invalid webhook signature")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "EVENT_RECEIVED")

	go h.processPayload(body)
}

func (h *WebhookHandler) validateSignature(body []byte, signature string) bool {
	if h.AppSecret == "" {
		return false
	}
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.AppSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

func (h *WebhookHandler) processPayload(body []byte) {
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Failed to parse webhook payload: %v", err)
		return
	}

	for _, entry := range payload.Entry {
		for _, m := range entry.Messaging {
			if m.Message == nil || m.Message.Text == "" {
				continue
			}

			var msg *Message
			if m.Message.IsEcho {
				msg = &Message{
					ID:         m.Message.MID,
					PSID:       m.Recipient.ID,
					PageID:     h.PageID,
					Text:       m.Message.Text,
					Direction:  "out",
					ReceivedAt: time.Unix(m.Timestamp/1000, 0),
				}
			} else {
				msg = &Message{
					ID:         m.Message.MID,
					PSID:       m.Sender.ID,
					PageID:     h.PageID,
					Text:       m.Message.Text,
					Direction:  "in",
					ReceivedAt: time.Unix(m.Timestamp/1000, 0),
				}
			}

			if h.Store != nil {
				if h.Store.MessageExists(msg.ID) {
					continue
				}
				if err := h.Store.SaveMessage(msg); err != nil {
					log.Printf("Failed to save message: %v", err)
				}
			}

			if m.Message.IsEcho {
				log.Printf("Echo message to %s: %s", msg.PSID, msg.Text)
				continue
			}

			log.Printf("Message from %s: %s", msg.PSID, msg.Text)

			// Feed inbound messages to debouncer for auto-reply
			if h.Debouncer != nil {
				h.Debouncer.Add(msg.PSID, DebouncerMessage{
					ID:   msg.ID,
					Text: msg.Text,
				})
			}
		}
	}
}

