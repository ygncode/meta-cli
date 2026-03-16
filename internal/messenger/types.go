package messenger

import "time"

type Message struct {
	ID          string    `json:"id"`
	PSID        string    `json:"psid"`
	PageID      string    `json:"page_id"`
	Text        string    `json:"text"`
	Direction   string    `json:"direction"`
	AutoReplied bool      `json:"auto_replied"`
	ReceivedAt  time.Time `json:"received_at"`
}

type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID        string      `json:"id"`
	Time      int64       `json:"time"`
	Messaging []Messaging `json:"messaging"`
}

type Messaging struct {
	Sender    Participant  `json:"sender"`
	Recipient Participant  `json:"recipient"`
	Timestamp int64        `json:"timestamp"`
	Message   *MsgPayload  `json:"message,omitempty"`
	Postback  *Postback    `json:"postback,omitempty"`
	Read      *ReadReceipt `json:"read,omitempty"`
}

type Participant struct {
	ID string `json:"id"`
}

type MsgPayload struct {
	MID    string `json:"mid"`
	Text   string `json:"text"`
	IsEcho bool   `json:"is_echo,omitempty"`
}

type Postback struct {
	Title   string `json:"title"`
	Payload string `json:"payload"`
}

type ReadReceipt struct {
	Watermark int64 `json:"watermark"`
}

type Conversation struct {
	ID           string `json:"id"`
	Participants string `json:"participants"`
	UpdatedTime  string `json:"updated_time"`
	MessageCount int    `json:"message_count"`
}
