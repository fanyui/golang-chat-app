package main

import (
	"log"
	"net/http"

	"ws.com/chat/internal/handlers"
)

func main() {
	mux := routes()
	log.Println("starting channels listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting server on port 8080")
	_ = http.ListenAndServe(":8080", mux)
}
