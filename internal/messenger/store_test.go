package messenger_test

import (
	"testing"
	"time"

	"github.com/ygncode/meta-cli/internal/messenger"
)

func openTestStore(t *testing.T) *messenger.Store {
	t.Helper()
	store, err := messenger.OpenStore(":memory:")
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestStoreSaveAndList(t *testing.T) {
	store := openTestStore(t)

	now := time.Now().Truncate(time.Second)
	msg := &messenger.Message{
		ID:         "mid_001",
		PSID:       "user_1",
		PageID:     "page_1",
		Text:       "hello",
		Direction:  "in",
		ReceivedAt: now,
	}

	if err := store.SaveMessage(msg); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	msgs, err := store.ListMessages("page_1", 10)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].ID != "mid_001" {
		t.Errorf("expected mid_001, got %s", msgs[0].ID)
	}
	if msgs[0].Text != "hello" {
		t.Errorf("expected hello, got %s", msgs[0].Text)
	}
	if msgs[0].Direction != "in" {
		t.Errorf("expected in, got %s", msgs[0].Direction)
	}
}

func TestStoreListByPage(t *testing.T) {
	store := openTestStore(t)

	now := time.Now()
	store.SaveMessage(&messenger.Message{ID: "1", PSID: "u1", PageID: "page_a", Text: "a", Direction: "in", ReceivedAt: now})
	store.SaveMessage(&messenger.Message{ID: "2", PSID: "u2", PageID: "page_b", Text: "b", Direction: "in", ReceivedAt: now})
	store.SaveMessage(&messenger.Message{ID: "3", PSID: "u3", PageID: "page_a", Text: "c", Direction: "in", ReceivedAt: now})

	msgs, err := store.ListMessages("page_a", 10)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages for page_a, got %d", len(msgs))
	}
}

func TestStoreListLimit(t *testing.T) {
	store := openTestStore(t)

	now := time.Now()
	for i := 0; i < 10; i++ {
		store.SaveMessage(&messenger.Message{
			ID: string(rune('a' + i)), PSID: "u", PageID: "p",
			Text: "msg", Direction: "in", ReceivedAt: now.Add(time.Duration(i) * time.Second),
		})
	}

	msgs, err := store.ListMessages("p", 3)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("expected 3 messages with limit, got %d", len(msgs))
	}
}

func TestStoreMessageExists(t *testing.T) {
	store := openTestStore(t)

	store.SaveMessage(&messenger.Message{
		ID: "mid_exists", PSID: "u", PageID: "p", Text: "t", Direction: "in", ReceivedAt: time.Now(),
	})

	if !store.MessageExists("mid_exists") {
		t.Error("expected MessageExists to return true")
	}
	if store.MessageExists("mid_nope") {
		t.Error("expected MessageExists to return false for missing ID")
	}
}

func TestStoreDuplicateInsert(t *testing.T) {
	store := openTestStore(t)

	msg := &messenger.Message{
		ID: "dup_1", PSID: "u", PageID: "p", Text: "first", Direction: "in", ReceivedAt: time.Now(),
	}
	store.SaveMessage(msg)

	// INSERT OR IGNORE should silently skip
	msg2 := &messenger.Message{
		ID: "dup_1", PSID: "u", PageID: "p", Text: "second", Direction: "in", ReceivedAt: time.Now(),
	}
	if err := store.SaveMessage(msg2); err != nil {
		t.Fatalf("SaveMessage duplicate: %v", err)
	}

	msgs, _ := store.ListMessages("p", 10)
	if len(msgs) != 1 {
		t.Errorf("expected 1 message after duplicate, got %d", len(msgs))
	}
	if msgs[0].Text != "first" {
		t.Errorf("expected original text preserved, got %s", msgs[0].Text)
	}
}

func TestStoreMarkAutoReplied(t *testing.T) {
	store := openTestStore(t)

	store.SaveMessage(&messenger.Message{
		ID: "mark_1", PSID: "u", PageID: "p", Text: "t", Direction: "in", ReceivedAt: time.Now(),
	})

	if err := store.MarkAutoReplied("mark_1"); err != nil {
		t.Fatalf("MarkAutoReplied: %v", err)
	}

	msgs, _ := store.ListMessages("p", 10)
	if !msgs[0].AutoReplied {
		t.Error("expected AutoReplied to be true after marking")
	}
}

func TestStoreAutoRepliedField(t *testing.T) {
	store := openTestStore(t)

	store.SaveMessage(&messenger.Message{
		ID: "ar_1", PSID: "u", PageID: "p", Text: "t",
		Direction: "out", AutoReplied: true, ReceivedAt: time.Now(),
	})

	msgs, _ := store.ListMessages("p", 10)
	if !msgs[0].AutoReplied {
		t.Error("expected AutoReplied to be true when saved as true")
	}
}
