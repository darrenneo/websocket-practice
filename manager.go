package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin:     checkOrigin,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Manager struct {
	clients ClientList
	sync.RWMutex

	OTP RetentionMap

	handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		OTP:      NewRetentionMap(ctx, 5*time.Second),
	}

	m.setupEventHandlers()
	return m
}

func (m *Manager) startCurr() {
	for _, currency := range currencies {
		go func(curr Currency) {
			for {
				nextJSON, err := curr.GetNextJSON()
				if err != nil {
					return
				}

				for client := range m.clients {
					if _, ok := client.subscribedCurrencies[curr.Name]; ok {
						client.egress <- Event{
							Type:    EventNewMessage,
							Payload: nextJSON,
						}
					}
				}

				time.Sleep(curr.Interval)
			}
		}(currency)
	}
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
	m.handlers[EventSubScribe] = Subscribe
	m.handlers[EventUnsubScribe] = Unsubscribe
}

func Unsubscribe(event Event, c *Client) error {
	var chatevent SubscribeEvent

	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("failed to unmarshal payload in unsubscribe: %w", err)
	}

	delete(c.subscribedCurrencies, chatevent.Currency)

	return nil
}

func Subscribe(event Event, c *Client) error {
	var chatevent SubscribeEvent

	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("failed to unmarshal payload in subscribe: %w", err)
	}

	c.subscribedCurrencies[chatevent.Currency] = true
	return nil
}

func SendMessage(event Event, c *Client) error {
	var chatevent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var broadMessage NewMessageEvent

	broadMessage.Sent = time.Now()
	broadMessage.Message = chatevent.Message
	broadMessage.From = chatevent.From

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast: %w", err)
	}

	outgoingEvent := Event{
		Payload: data,
		Type:    EventNewMessage,
	}

	for client := range c.manager.clients {
		client.egress <- outgoingEvent
	}

	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("there is no such event type")
	}
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	// otp := r.URL.Query().Get("otp")
	// if otp == "" {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	return
	// }

	// if !m.OTP.VerifyOTP(otp) {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	return
	// }

	log.Println("Websocket connection established")

	// Upgrade the HTTP connection to a websocket connection
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)

	m.addClient(client)

	// Start Client Process
	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	type userLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req userLoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "test" && req.Password == "123" {
		type response struct {
			OTP string `json:"otp"`
		}

		otp := m.OTP.NewOTP()

		resp := response{
			OTP: otp.Key,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	fmt.Println(origin)

	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false
	}
}
