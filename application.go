package main

import (
	"log"
	"net/http"

	"ws.com/chat/internal/handlers"
	"ws.com/chat/routes"
)

func main() {
	mux := routes.Routes()
	log.Println("starting channels listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting server on port 8080")
	_ = http.ListenAndServe(":8080", mux)
}
