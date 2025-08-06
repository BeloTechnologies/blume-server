package socket

import (
	"blume-server/game"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var zapLogger *zap.Logger

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	Clients   map[string]*websocket.Conn
	Positions map[string]string
	GameState *game.GameState
	mu        sync.Mutex
}

func NewServer(logger *zap.Logger, gameState *game.GameState) *Server {
	zapLogger = logger
	return &Server{
		Clients:   make(map[string]*websocket.Conn),
		Positions: make(map[string]string),
		GameState: gameState,
	}
}

func (s *Server) broadcast(senderID string, message []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, conn := range s.Clients {
		if id != senderID {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				zapLogger.Error("Error sending to client", zap.String("id", id), zap.Error(err))
			}
		}
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zapLogger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	playerID := uuid.New().String()
	zapLogger.Info("Client connected", zap.String("id", playerID))

	s.mu.Lock()
	s.Clients[playerID] = conn
	s.GameState.AddPlayer(&game.Player{ID: playerID})
	s.mu.Unlock()

	// Send assigned ID
	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"id", "id":"`+playerID+`"}`))
	if err != nil {
		zapLogger.Error("Error sending ID", zap.Error(err))
		return
	}

	// Send positions of other players to this new client
	s.mu.Lock()
	for id, pos := range s.Positions {
		if id != playerID {
			conn.WriteMessage(websocket.TextMessage, []byte(pos))
		}
	}
	s.mu.Unlock()

	// Handle messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			zapLogger.Error("Error reading message", zap.String("id", playerID), zap.Error(err))
			break
		}

		zapLogger.Info("Received message", zap.String("id", playerID), zap.ByteString("message", message))

		// Save last position if it's a move message
		if string(message[10:15]) == "move\"" {
			s.mu.Lock()
			s.Positions[playerID] = string(message)
			s.mu.Unlock()
		}

		s.broadcast(playerID, message)
	}

	// Clean up on disconnect
	zapLogger.Info("Client disconnected", zap.String("id", playerID))
	s.mu.Lock()
	delete(s.Clients, playerID)
	delete(s.Positions, playerID)
	s.GameState.RemovePlayer(playerID)
	s.mu.Unlock()

	// Notify others that this player left
	leaveMsg := []byte(`{"type":"leave", "id":"` + playerID + `"}`)
	s.broadcast(playerID, leaveMsg)
}
