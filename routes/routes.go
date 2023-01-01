package routes

import (
	"net/http"

	"github.com/bmizerany/pat"
	"ws.com/chat/internal/handlers"
)

func Routes() http.Handler {
	mux := pat.New()
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))
	return mux
}
