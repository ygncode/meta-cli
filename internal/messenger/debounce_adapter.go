package messenger

import "github.com/ygncode/meta-cli/internal/debounce"

// DebouncerAdapter wraps debounce.Debouncer to satisfy DebouncerInterface.
type DebouncerAdapter struct {
	D *debounce.Debouncer
}

func (a *DebouncerAdapter) Add(psid string, msg DebouncerMessage) {
	a.D.Add(psid, debounce.Message{
		ID:   msg.ID,
		Text: msg.Text,
	})
}
