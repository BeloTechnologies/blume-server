package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[string]*websocket.Conn)
var positions = make(map[string]string)
var mu sync.Mutex

func broadcast(senderID string, message []byte) {
	mu.Lock()
	defer mu.Unlock()
	for id, conn := range clients {
		if id != senderID {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Error sending to client %s: %v", id, err)
			}
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	playerID := uuid.New().String()
	log.Printf("Client connected: %s", playerID)

	mu.Lock()
	clients[playerID] = conn
	mu.Unlock()

	// Send assigned ID
	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"id", "id":"`+playerID+`"}`))
	if err != nil {
		log.Println("Error sending ID:", err)
		return
	}

	// Send positions of other players to this new client
	mu.Lock()
	for id, pos := range positions {
		if id != playerID {
			conn.WriteMessage(websocket.TextMessage, []byte(pos))
		}
	}
	mu.Unlock()

	// Handle messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client %s disconnected", playerID)
			break
		}

		log.Printf("Received from %s: %s", playerID, message)

		// Save last position if it's a move message
		if string(message[10:15]) == "move\"" {
			mu.Lock()
			positions[playerID] = string(message)
			mu.Unlock()
		}

		broadcast(playerID, message)
	}

	// Clean up on disconnect
	mu.Lock()
	delete(clients, playerID)
	delete(positions, playerID)
	mu.Unlock()

	// Notify others that this player left
	leaveMsg := []byte(`{"type":"leave", "id":"` + playerID + `"}`)
	broadcast(playerID, leaveMsg)
}
func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Starting Blume server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
