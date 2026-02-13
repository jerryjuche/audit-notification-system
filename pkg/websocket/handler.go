package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader configures the WebSocket upgrade process.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev; restrict in prod.
	},
}

// EchoHandler upgrades HTTP to WS and echoes messages.
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer conn.Close() // Close connection when function ends.

	log.Println("Client connected")

	for {
		// Read message from client.
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}

		// Echo back to client.
		if err := conn.WriteMessage(msgType, msg); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}
