package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)

// var clients = make(map[WebScoketConnection]string)
var clients = make(map[string]map[WebScoketConnection]string)

var updates = make(map[string][]Update)
var newlyJoined = make(map[string][]string)

// var updates = []Update{}
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebScoketConnection struct {
	*websocket.Conn
}

//	type  Update struct {
//		/// The changes made by this update.
//		changes ChangeSet,
//		/// The effects in this update. There'll only ever be effects here
//		/// when you configure your collab extension with a
//		/// [`sharedEffects`](#collab.collab^config.sharedEffects) option.
//		effects <any>[]StateEffect
//		/// The [ID](#collab.collab^config.clientID) of the client who
//		/// created this update.
//		clientID: string
//	  }
//
// WsJsonResponse defines the response sent back from websocket
type WsJsonResponse struct {
	Username       string   `json:"username"`
	Workspace      string   `json:"workspace"`
	Version        int      `json:"version"`
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	ConnectedUsers []string `json:"connected_users"`
	Payload        []Update `json:"payload"`
}

type Update struct {
	ClientID string `json:"client_id"`
	Changes  string `json:"changes"`
}
type WsPayload struct {
	Workspace string              `json:"workspace"`
	Version   int                 `json:"version"`
	Action    string              `json:"action"`
	Username  string              `json:"username"`
	Message   string              `json:"message"`
	Payload   []Update            `json:"payload"`
	Conn      WebScoketConnection `json:"-"`
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
	// clients[conn] = ""
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
		fmt.Println("some crasyz: ", e.Action)
		// response.Action = "Got here"
		// response.Message = fmt.Sprintf("Some message, and action was %s", e.Action)
		// broadCastToAll(response)
		switch e.Action {
		case "createworkspace":
			// get the current workspace if it is available in the first place

			clients[e.Workspace] = make(map[WebScoketConnection]string)
			updates[e.Workspace] = []Update{}
			response.Action = "success"
			response.Workspace = e.Workspace
			response.Username = e.Username
			e.Conn.WriteJSON(response)

		case "join_workspace":
			// get the current workspace if it is available in the first place
			workspace, ok := clients[e.Workspace]
			if !ok {
				// the workspace doesn't exist send this user a warning
				response.Action = "error"
				response.Message = "Workspace doesn't exist"
				e.Conn.WriteJSON(response)
			} else {
				workspace[e.Conn] = e.Username
				response.Action = e.Action
				response.Workspace = e.Workspace
				response.Username = e.Username
				response.Message = "Workspace joined successfull"
				e.Conn.WriteJSON(response)
			}

		case "getDocument":

			response.Workspace = e.Workspace
			response.Username = "harisu"
			// response.Action = "getDocument"
			// response.Version = 0

			// e.Conn.WriteJSON(response)

			response.Action = "pushUpdates"
			var workspaceupdates = updates[e.Workspace]
			response.Payload = workspaceupdates
			response.Version = len(workspaceupdates)
			e.Conn.WriteJSON(response)
		case "pushUpdates":
			fmt.Printf("event payload: %v", e.Payload)
			// if e.Version == len(updates) {
			var update = e.Payload
			workspaceupdates := updates[e.Workspace]

			updates[e.Workspace] = append(workspaceupdates, update...)
			response.Action = "pushUpdates"
			response.Workspace = e.Workspace
			response.Payload = e.Payload
			response.Version = len(updates[e.Workspace])
			response.Username = e.Username
			// e.Conn.WriteJSON(response)
			broadCastToAll(e.Workspace, response)

			// }
			// if (e.version != len(updates)) {
			//   resp(false)
			// } else {
			//   for (let update of data.updates) {
			// 	// Convert the JSON representation to an actual ChangeSet
			// 	// instance
			// 	let changes = ChangeSet.fromJSON(update.changes)
			// 	updates.push({changes, clientID: update.clientID})
			// 	doc = changes.apply(doc)
			//   }
			//   resp(true)
			//   // Notify pending requests
			//   while (pending.length) pending.pop()!(data.updates)
			// }
			//   }
		case "pullUpdates":
			response.Action = "pullUpdates"
			response.Workspace = e.Workspace
			response.Payload = e.Payload
			response.Username = "harisu"
			e.Conn.WriteJSON(response)
		}
	}
}

//	func removeItemFromSlice(s []string, i int) []string {
//		s[i] = s[len(s)-1]
//		return s[:len(s)-1]
//	}
func sendToUser(workspaceId string, username string, response WsJsonResponse) error {
	for conn, x := range clients[workspaceId] {
		if x != "" && x == username {
			response.Username = x
			err := conn.WriteJSON(response)
			if err != nil {
				log.Println("websocket erro")
				_ = conn.Close()
				delete(clients[workspaceId], conn)
			}
			return nil
		}
	}
	return fmt.Errorf("user not found")
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
			//do nothing for now but remember to handle this error
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}
func broadCastToAll(workspaceId string, response WsJsonResponse) {
	for client, username := range clients[workspaceId] {
		response.Username = username
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket erro")
			_ = client.Close()
			delete(clients[workspaceId], client)
		}
	}
}

// Todo: Build the receiver UI to get the exchanged messages.
//
