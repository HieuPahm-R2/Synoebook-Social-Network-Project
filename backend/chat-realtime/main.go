package main

import (
	"log"
	"os"
	"realTimeChat/realtime"
)

func main() {
	manager := realtime.NewConnectionManager(realtime.GetUserFriends)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5002"
	}
	if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("Starting Realtime Chat WebSocket server on port %s...", port)
	if err := realtime.StartWebSocketServer(port, manager); err != nil {
		log.Fatalf("Failed to start WebSocket server: %v", err)
	}
}
