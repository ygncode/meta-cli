package debounce

import (
	"sync"
	"time"
)

// Message represents a debounced message with an ID and text.
type Message struct {
	ID   string
	Text string
}

// Callback is called when the debounce window expires for a PSID.
// It receives the PSID and all messages collected during the window.
type Callback func(psid string, messages []Message)

type pending struct {
	timer    *time.Timer
	messages []Message
}

// Debouncer collects messages per PSID and fires a callback after a
// configurable quiet period (no new messages within the window).
type Debouncer struct {
	window  time.Duration
	cb      Callback
	mu      sync.Mutex
	pending map[string]*pending
	stopped bool
}

// New creates a new Debouncer with the given window duration and callback.
func New(window time.Duration, cb Callback) *Debouncer {
	return &Debouncer{
		window:  window,
		cb:      cb,
		pending: make(map[string]*pending),
	}
}

// Add adds a message for the given PSID. If a timer is already running
// for this PSID, it is reset. When the timer fires (no new messages
// within the window), the callback is invoked with all collected messages.
func (d *Debouncer) Add(psid string, msg Message) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stopped {
		return
	}

	p, ok := d.pending[psid]
	if ok {
		p.timer.Stop()
		p.messages = append(p.messages, msg)
	} else {
		p = &pending{
			messages: []Message{msg},
		}
		d.pending[psid] = p
	}

	p.timer = time.AfterFunc(d.window, func() {
		d.fire(psid)
	})
}

func (d *Debouncer) fire(psid string) {
	d.mu.Lock()
	p, ok := d.pending[psid]
	if !ok {
		d.mu.Unlock()
		return
	}
	msgs := p.messages
	delete(d.pending, psid)
	d.mu.Unlock()

	d.cb(psid, msgs)
}

// Stop cancels all pending timers and prevents future Add calls from
// scheduling new timers. It is safe to call multiple times.
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.stopped = true
	for psid, p := range d.pending {
		p.timer.Stop()
		delete(d.pending, psid)
	}
}
