package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected")

	err = conn.WriteMessage(websocket.TextMessage, []byte("hello world"))
	if err != nil {
		log.Println("Error writing message:", err)
		return
	}

	for {
		// Read messages from the WebSocket connection
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Printf("Received message: %s", message)

		// Echo the message back to the client
		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Starting Blume server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
