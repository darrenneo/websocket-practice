package test

// import (
// 	"encoding/json"
// 	"fmt"
// 	"math/rand"
// 	"strconv"
// 	"time"

// 	"github.com/gorilla/websocket"
// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// )

// type Currency struct {
// 	Name           string        `json:"name"`
// 	Value          int64         `json:"-"`
// 	ValueString    string        `json:"value"`
// 	Decimal        uint8         `json:"decimal"`
// 	ChangeMin      int64         `json:"-"`
// 	ChangeMax      int64         `json:"-"`
// 	Interval       time.Duration `json:"-"`
// 	IntervalString int           `json:"interval_in_ms"`
// }

// func (c Currency) GetNextJSON() ([]byte, error) {
// 	newChange := randRange(c.ChangeMin, c.ChangeMax)
// 	// Currency shall not go negative
// 	if newChange < 0 && c.Value+newChange <= 0 {
// 		c.Value = c.Value - newChange
// 	} else {
// 		c.Value = c.Value + newChange
// 	}
// 	c.ValueString = strconv.FormatInt(c.Value, 10)
// 	return json.Marshal(c)
// }

// // min is inclusive; max is exclusive
// func randRange(min, max int64) int64 {
// 	return rand.Int63n(max-min) + min
// }

// var currencies = []Currency{
// 	{
// 		Name:           "Nani",
// 		Value:          123456789012345678,
// 		Decimal:        8,
// 		ChangeMin:      -12345678,
// 		ChangeMax:      12345678,
// 		Interval:       1 * time.Second,
// 		IntervalString: 1000,
// 	},
// 	{
// 		Name:           "Programming",
// 		Value:          999888777666,
// 		Decimal:        6,
// 		ChangeMin:      -555444333,
// 		ChangeMax:      555444333,
// 		Interval:       5 * time.Second,
// 		IntervalString: 5000,
// 	},
// 	{
// 		Name:           "Is",
// 		Value:          987654321098765432,
// 		Decimal:        12,
// 		ChangeMin:      -12345,
// 		ChangeMax:      12345,
// 		Interval:       800 * time.Millisecond,
// 		IntervalString: 800,
// 	},
// 	{
// 		Name:           "Fun",
// 		Value:          11111111111,
// 		Decimal:        11,
// 		ChangeMin:      -1111,
// 		ChangeMax:      1111,
// 		Interval:       1111 * time.Millisecond,
// 		IntervalString: 1111,
// 	},
// }
// var upgrader = websocket.Upgrader{}

// func ws(c echo.Context) error {
// 	fmt.Println("New socket")
// 	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer ws.Close()

// 	doneCh := make(chan interface{})
// 	msgCh := make(chan []byte)
// 	closeCh := make(chan interface{})

// 	go func() { // GoRoutine 1
// 		for {
// 			// Read
// 			msgType, _, err := ws.ReadMessage()
// 			if err != nil {
// 				fmt.Println("Closing socket due to error")
// 				closeCh <- nil
// 				return
// 			}

// 			if msgType == websocket.CloseMessage {
// 				fmt.Println("Closing socket due to client closing")
// 				closeCh <- nil
// 				return
// 			}
// 		}
// 	}()

// 	go func(ch <-chan []byte, done chan<- interface{}) { // GoRoutine 2
// 		for {
// 			select {
// 			case msg := <-ch:
// 				err := ws.WriteMessage(websocket.TextMessage, msg)
// 				if err != nil {
// 					done <- nil
// 					return
// 				}
// 			case <-closeCh:
// 				done <- nil
// 				return
// 			}
// 		}
// 	}(msgCh, doneCh)

// 	for _, currency := range currencies { // GoRoutine Group 3
// 		go func(curr Currency, ch chan<- []byte) {
// 			for {
// 				nextJSON, err := curr.GetNextJSON()
// 				if err != nil {
// 					return
// 				}

// 				ch <- nextJSON

// 				time.Sleep(curr.Interval)
// 			}
// 		}(currency, msgCh)
// 	}

// 	<-doneCh

// 	return nil
// }

// func mains() {
// 	e := echo.New()
// 	e.Use(middleware.Logger())
// 	e.Use(middleware.Recover())
// 	e.GET("/ws", ws)
// 	e.Logger.Fatal(e.Start(":5050"))
// }
