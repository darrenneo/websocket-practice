package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait = 10 * time.Second

	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	//
	connection *websocket.Conn
	manager    *Manager

	// egress is used to avoid concurrent writes to the websocket connection
	egress chan Event

	subscribedCurrencies map[string]bool
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection:           conn,
		manager:              manager,
		egress:               make(chan Event),
		subscribedCurrencies: make(map[string]bool),
	}
}

//! Moved to manager
// func (c *Client) start() {
// 	for _, currency := range currencies {
// 		go func(curr Currency, ch chan<- Event) {
// 			for {
// 				nextJSON, err := curr.GetNextJSON()
// 				if err != nil {
// 					return
// 				}
// 				c.egress <- Event{
// 					Type:    EventNewMessage,
// 					Payload: nextJSON,
// 				}

// 				time.Sleep(curr.Interval)

// 			}
// 		}(currency, c.egress)
// 	}
// }

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

	ticker := time.NewTicker(pingInterval)

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

		case <-ticker.C:
			//! Ping
			// log.Println("ping")

			// send ping message
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Println("failed to send ping message: ", err)
				return
			}

			// if err := c.connection.WriteMessage(websocket.TextMessage, []byte("Data from server")); err != nil {
			// 	log.Println("failed to send data from server: ", err)
			// 	return
			// }

		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	//! Pong
	// log.Println("pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

// func generateCurr(c *Client, curren Currency) {
// 	for _, currency := range currencies {
// 		go func(curr Currency, ch chan<- []byte) {
// 			for {
// 				nextJSON, err := currency.GetNextJSON()
// 				if err != nil {
// 					return
// 				}
// 				c.egress <- Event{
// 					Type:    EventNewMessage,
// 					Payload: nextJSON,
// 				}

// 				time.Sleep(currency.Interval)

// 			}
// 		}(curren, c.egress)
// 	}
// }
