package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"websocket-practice/internal/settings"

	"github.com/gorilla/websocket"
)

var (
	pongWait     time.Duration
	ackInterval  time.Duration
	pingInterval time.Duration
)

type ClientList map[*Client]bool

type Client struct {
	//
	connection *websocket.Conn
	manager    *Manager

	// egress is used to avoid concurrent writes to the websocket connection
	egress chan Event

	subscribedCurrencies map[string]bool

	acknowledge bool

	sync.RWMutex
}

func InitTimers() {
	pongWait = time.Duration(settings.Get().Configs.PongWait) * time.Second
	ackInterval = time.Duration(settings.Get().Configs.AckInterval) * time.Second
	pingInterval = (pongWait * 9) / 10
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection:           conn,
		manager:              manager,
		egress:               make(chan Event),
		subscribedCurrencies: make(map[string]bool),
		acknowledge:          false,
	}
}

func (c *Client) readMessages() {
	defer func() {
		// Clean up connection
		c.manager.removeClient(c)
	}()

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("failed to set read deadline: ", err)
		return
	}

	c.connection.SetReadLimit(512)

	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, message, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
			break
		}

		var request Event

		if err := json.Unmarshal(message, &request); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			continue
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Printf("failed to route event: %v", err)
		}

	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	pingTicker := time.NewTicker(pingInterval)

	ackTicker := time.NewTicker(ackInterval)
	defer ackTicker.Stop()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connetion closed: ", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("failed to marshal message: %v", err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to send message: %v", err)
			}

		case <-pingTicker.C:
			//! Ping
			// send ping message
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Println("failed to send ping message: ", err)
				return
			}
			if !c.acknowledge {

				pingEvent := Event{
					Type:    EventPing,
					Payload: json.RawMessage(`{"message": "ping"}`),
				}

				pingMessage, err := json.Marshal(pingEvent)
				if err != nil {
					log.Printf("failed to marshal message: %v", err)
					return
				}

				if err := c.connection.WriteMessage(websocket.TextMessage, pingMessage); err != nil {
					log.Println("failed to send ping message: ", err)
					return
				}
			}

		case <-ackTicker.C:
			//? Acknowledge
			ackEvent := Event{
				Type:    EventAcknowledge,
				Payload: json.RawMessage(`{"message": "client did not acknowledge, disconnecting"}`),
			}

			ackMessage, err := json.Marshal(ackEvent)
			if err != nil {
				log.Printf("failed to marshal message: %v", err)
				return
			}

			if !c.acknowledge {
				if err := c.connection.WriteMessage(websocket.TextMessage, ackMessage); err != nil {
					log.Println("failed to send ping message: ", err)
					return
				}
				log.Println("Client did not acknowledge")
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	//! Pong
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) checkSubscribedCurrency(curr string) bool {
	if _, ok := c.subscribedCurrencies[curr]; !ok {
		return false
	}

	return true
}
