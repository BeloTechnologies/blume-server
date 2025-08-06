package main

import (
	"blume-server/socket"
	"log"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	var err error
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Cannot initialize zap logger: %v", err)
	}
	defer zapLogger.Sync() // flushes buffer, if any

	// Init echo
	e := echo.New()

	// Echo middleware using zap
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			zapLogger.Info("HTTP request",
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", res.Status),
			)
			return next(c)
		}
	})

	// Initialize WebSocket server
	wsServer := socket.NewServer(zapLogger)

	e.GET("/ws", func(c echo.Context) error {
		wsServer.HandleWebSocket(c.Response().Writer, c.Request())
		return nil
	})

	zapLogger.Info("Starting server on :8080")

	// Start server
	if err := e.Start(":8080"); err != nil {
		zapLogger.Fatal("Failed to start server", zap.Error(err))
	}
	zapLogger.Info("Server stopped")
}
