package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	setupAPI()
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func setupAPI() {
	ctx := context.Background()

	manager := NewManager(ctx)

	go manager.startCurr()

	http.HandleFunc("/ws", manager.serveWS)
	http.HandleFunc("/login", manager.loginHandler)
}
