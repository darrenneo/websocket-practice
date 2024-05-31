package main

import (
	"context"
	"log"
	"net/http"

	"websocket-practice/internal/settings"
	"websocket-practice/internal/websocket"
)

func main() {
	settings.LoadSettings()

	setupAPI()

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func setupAPI() {
	ctx := context.Background()

	manager := websocket.NewManager(ctx)

	go manager.StartCurr()

	http.HandleFunc("/ws", manager.ServeWS)
	http.HandleFunc("/login", manager.LoginHandler)
}
