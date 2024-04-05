package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Connection struct {
	WsConn *websocket.Conn
	Egress chan []byte
}

func NewConnection(ws *websocket.Conn, egrees chan []byte) *Connection {
	return &Connection{
		WsConn: ws,
		Egress: egrees,
	}
}

var connections = make(map[string]*Connection)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

func HandleWs(w http.ResponseWriter, r *http.Request) {
	// accept all the connections to avoid CORS errors for now
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	newConnection := &Connection{WsConn: ws, Egress: make(chan []byte)}
	connections[ws.RemoteAddr().String()] = newConnection 

	log.Println("a new connection made:", ws.RemoteAddr().String())
	go ManageWsConn(newConnection)
}

func ManageWsConn(conn *Connection) {
	ws := conn.WsConn 
	// remove the connection from the in-memory storage if the connection ended
	defer func() {
		delete(connections, ws.RemoteAddr().String())
	}()

	for {
		msgType, buffer, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
			// whatever else is not really an error. would be normal closure
			break
		}

		log.Printf("message type: %d", msgType)
		log.Println(string(buffer))

		// connection can only write one message at a time
		// we
		if err := ws.WriteMessage(msgType, []byte("your message was received!")); err != nil {
			log.Println(err)
		}
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ws", HandleWs)
	log.Println("listening to port 8000")
	http.ListenAndServe(":8000", mux)
}

// func sendHello(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("hello client!"))
// }
