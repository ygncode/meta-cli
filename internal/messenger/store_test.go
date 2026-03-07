package messenger_test

import (
	"fmt"
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

func TestRecentMessages(t *testing.T) {
	store := openTestStore(t)

	now := time.Now().Truncate(time.Second)
	store.SaveMessage(&messenger.Message{ID: "1", PSID: "user_1", PageID: "page_1", Text: "hello", Direction: "in", ReceivedAt: now})
	store.SaveMessage(&messenger.Message{ID: "2", PSID: "user_2", PageID: "page_1", Text: "other", Direction: "in", ReceivedAt: now})
	store.SaveMessage(&messenger.Message{ID: "3", PSID: "user_1", PageID: "page_2", Text: "diff page", Direction: "in", ReceivedAt: now})

	msgs, err := store.RecentMessages("page_1", "user_1", 10)
	if err != nil {
		t.Fatalf("RecentMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message for user_1/page_1, got %d", len(msgs))
	}
	if msgs[0].Text != "hello" {
		t.Errorf("expected 'hello', got %s", msgs[0].Text)
	}
}

func TestRecentMessagesOrder(t *testing.T) {
	store := openTestStore(t)

	now := time.Now().Truncate(time.Second)
	store.SaveMessage(&messenger.Message{ID: "1", PSID: "u1", PageID: "p1", Text: "first", Direction: "in", ReceivedAt: now})
	store.SaveMessage(&messenger.Message{ID: "2", PSID: "u1", PageID: "p1", Text: "second", Direction: "out", ReceivedAt: now.Add(time.Second)})
	store.SaveMessage(&messenger.Message{ID: "3", PSID: "u1", PageID: "p1", Text: "third", Direction: "in", ReceivedAt: now.Add(2 * time.Second)})

	msgs, err := store.RecentMessages("p1", "u1", 10)
	if err != nil {
		t.Fatalf("RecentMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	// Should be oldest first
	if msgs[0].Text != "first" {
		t.Errorf("expected 'first' at index 0, got %s", msgs[0].Text)
	}
	if msgs[1].Text != "second" {
		t.Errorf("expected 'second' at index 1, got %s", msgs[1].Text)
	}
	if msgs[2].Text != "third" {
		t.Errorf("expected 'third' at index 2, got %s", msgs[2].Text)
	}
}

func TestRecentMessagesLimit(t *testing.T) {
	store := openTestStore(t)

	now := time.Now().Truncate(time.Second)
	for i := 0; i < 10; i++ {
		store.SaveMessage(&messenger.Message{
			ID: fmt.Sprintf("msg_%d", i), PSID: "u1", PageID: "p1",
			Text: fmt.Sprintf("msg %d", i), Direction: "in",
			ReceivedAt: now.Add(time.Duration(i) * time.Second),
		})
	}

	msgs, err := store.RecentMessages("p1", "u1", 3)
	if err != nil {
		t.Fatalf("RecentMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages with limit, got %d", len(msgs))
	}
	// Should return the 3 most recent, in chronological order
	if msgs[0].Text != "msg 7" {
		t.Errorf("expected 'msg 7', got %s", msgs[0].Text)
	}
	if msgs[2].Text != "msg 9" {
		t.Errorf("expected 'msg 9', got %s", msgs[2].Text)
	}
}

func TestRecentMessagesBothDirections(t *testing.T) {
	store := openTestStore(t)

	now := time.Now().Truncate(time.Second)
	store.SaveMessage(&messenger.Message{ID: "1", PSID: "u1", PageID: "p1", Text: "question", Direction: "in", ReceivedAt: now})
	store.SaveMessage(&messenger.Message{ID: "2", PSID: "u1", PageID: "p1", Text: "answer", Direction: "out", ReceivedAt: now.Add(time.Second)})

	msgs, err := store.RecentMessages("p1", "u1", 10)
	if err != nil {
		t.Fatalf("RecentMessages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Direction != "in" {
		t.Errorf("expected first message direction 'in', got %s", msgs[0].Direction)
	}
	if msgs[1].Direction != "out" {
		t.Errorf("expected second message direction 'out', got %s", msgs[1].Direction)
	}
}

func TestRecentMessagesEmpty(t *testing.T) {
	store := openTestStore(t)

	msgs, err := store.RecentMessages("p1", "unknown_user", 10)
	if err != nil {
		t.Fatalf("RecentMessages: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages for unknown PSID, got %d", len(msgs))
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
