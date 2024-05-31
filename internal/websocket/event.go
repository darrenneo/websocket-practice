package websocket

import (
	"encoding/json"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage    = "send_message"
	EventNewMessage     = "new_message"
	EventSubScribe      = "subscribe"
	EventUnsubScribe    = "unsubscribe"
	EventAcknowledge    = "acknowledge"
	EventConnectionInit = "connection_init"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type SubscribeEvent struct {
	Currency string `json:"currency"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}
