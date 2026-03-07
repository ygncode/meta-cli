package debounce_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ygncode/meta-cli/internal/debounce"
)

func TestSingleMessage(t *testing.T) {
	var mu sync.Mutex
	var got []debounce.Message
	var gotPSID string

	d := debounce.New(20*time.Millisecond, func(psid string, msgs []debounce.Message) {
		mu.Lock()
		gotPSID = psid
		got = msgs
		mu.Unlock()
	})
	defer d.Stop()

	d.Add("user_1", debounce.Message{ID: "1", Text: "hello"})

	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if gotPSID != "user_1" {
		t.Errorf("expected psid user_1, got %s", gotPSID)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 message, got %d", len(got))
	}
	if got[0].Text != "hello" {
		t.Errorf("expected 'hello', got %s", got[0].Text)
	}
}

func TestDebounceResets(t *testing.T) {
	var mu sync.Mutex
	var got []debounce.Message
	callCount := 0

	d := debounce.New(30*time.Millisecond, func(psid string, msgs []debounce.Message) {
		mu.Lock()
		got = msgs
		callCount++
		mu.Unlock()
	})
	defer d.Stop()

	d.Add("user_1", debounce.Message{ID: "1", Text: "msg1"})
	time.Sleep(10 * time.Millisecond)
	d.Add("user_1", debounce.Message{ID: "2", Text: "msg2"})
	time.Sleep(10 * time.Millisecond)
	d.Add("user_1", debounce.Message{ID: "3", Text: "msg3"})

	// Wait for debounce to fire
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 1 {
		t.Errorf("expected callback called once, got %d", callCount)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(got))
	}
}

func TestMultiplePSIDs(t *testing.T) {
	var mu sync.Mutex
	results := make(map[string][]debounce.Message)

	d := debounce.New(20*time.Millisecond, func(psid string, msgs []debounce.Message) {
		mu.Lock()
		results[psid] = msgs
		mu.Unlock()
	})
	defer d.Stop()

	d.Add("user_1", debounce.Message{ID: "1", Text: "from user 1"})
	d.Add("user_2", debounce.Message{ID: "2", Text: "from user 2"})

	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(results) != 2 {
		t.Fatalf("expected 2 PSIDs, got %d", len(results))
	}
	if len(results["user_1"]) != 1 || results["user_1"][0].Text != "from user 1" {
		t.Errorf("unexpected user_1 messages: %v", results["user_1"])
	}
	if len(results["user_2"]) != 1 || results["user_2"][0].Text != "from user 2" {
		t.Errorf("unexpected user_2 messages: %v", results["user_2"])
	}
}

func TestMessageOrder(t *testing.T) {
	var mu sync.Mutex
	var got []debounce.Message

	d := debounce.New(20*time.Millisecond, func(psid string, msgs []debounce.Message) {
		mu.Lock()
		got = msgs
		mu.Unlock()
	})
	defer d.Stop()

	d.Add("u1", debounce.Message{ID: "1", Text: "first"})
	d.Add("u1", debounce.Message{ID: "2", Text: "second"})
	d.Add("u1", debounce.Message{ID: "3", Text: "third"})

	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(got))
	}
	if got[0].Text != "first" || got[1].Text != "second" || got[2].Text != "third" {
		t.Errorf("unexpected order: %v", got)
	}
}

func TestStop(t *testing.T) {
	callCount := 0

	d := debounce.New(20*time.Millisecond, func(psid string, msgs []debounce.Message) {
		callCount++
	})

	d.Add("u1", debounce.Message{ID: "1", Text: "hello"})
	d.Stop()

	time.Sleep(60 * time.Millisecond)

	if callCount != 0 {
		t.Errorf("expected 0 callbacks after Stop, got %d", callCount)
	}
}

func TestStopIdempotent(t *testing.T) {
	d := debounce.New(20*time.Millisecond, func(psid string, msgs []debounce.Message) {})

	// Calling Stop multiple times should not panic
	d.Stop()
	d.Stop()
	d.Stop()
}

func TestAddAfterStop(t *testing.T) {
	callCount := 0

	d := debounce.New(20*time.Millisecond, func(psid string, msgs []debounce.Message) {
		callCount++
	})

	d.Stop()

	// Adding after Stop should not panic
	d.Add("u1", debounce.Message{ID: "1", Text: "hello"})

	time.Sleep(60 * time.Millisecond)

	if callCount != 0 {
		t.Errorf("expected 0 callbacks after Add-after-Stop, got %d", callCount)
	}
}

func TestZeroWindow(t *testing.T) {
	var mu sync.Mutex
	var got []debounce.Message

	d := debounce.New(0, func(psid string, msgs []debounce.Message) {
		mu.Lock()
		got = msgs
		mu.Unlock()
	})
	defer d.Stop()

	d.Add("u1", debounce.Message{ID: "1", Text: "immediate"})

	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 message with zero window, got %d", len(got))
	}
	if got[0].Text != "immediate" {
		t.Errorf("expected 'immediate', got %s", got[0].Text)
	}
}

func TestConcurrentAdds(t *testing.T) {
	var mu sync.Mutex
	var got []debounce.Message

	d := debounce.New(30*time.Millisecond, func(psid string, msgs []debounce.Message) {
		mu.Lock()
		got = msgs
		mu.Unlock()
	})
	defer d.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			d.Add("u1", debounce.Message{ID: string(rune('a' + i)), Text: "msg"})
		}(i)
	}
	wg.Wait()

	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 100 {
		t.Errorf("expected 100 messages from concurrent adds, got %d", len(got))
	}
}
