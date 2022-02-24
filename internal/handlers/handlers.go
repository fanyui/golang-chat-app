package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)
var clients = make(map[WebScoketConnection]string)
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WebScoketConnection struct {
	*websocket.Conn
}

//WsJsonResponse defines the response sent back from websocket
type WsJsonResponse struct {
	Action      string `json: "action"`
	Message     string `json: "message"`
	MessageType string `json: "message_type"`
}
type WsPayload struct {
	Action   string              `json: "action"`
	Username string              `json: "username"`
	Message  string              `json: "message"`
	Conn     WebScoketConnection `json: "-"`
}

// WsEndpoint upgreades a connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client connected to endpoint")
	var response WsJsonResponse
	response.Message = `<em><small> Connected to server</small></em>`
	conn := WebScoketConnection{Conn: ws}
	clients[conn] = ""
	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}
	go ListenForWs(&conn)
}

func ListenToWsChannel() {
	var response WsJsonResponse
	for {
		e := <-wsChan
		response.Action = "Got here"
		response.Message = fmt.Sprintf("Some message, and action was %s", e.Action)

	}
}
func broadCastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket erro")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func ListenForWs(conn *WebScoketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()
	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			//do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}
func renderPage(w http.ResponseWriter, templ string, data jet.VarMap) error {
	view, err := views.GetTemplate(templ)
	if err != nil {
		log.Println(err)
		return err
	}
	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
