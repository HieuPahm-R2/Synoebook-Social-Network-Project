package main

import (
	"log"
	"realTimeNotification/grpc"
	"realTimeNotification/realtime"
	"sync"

	"github.com/gofiber/websocket/v2"
)

func main() {

	wsMu := sync.Mutex{}
	ws := make(map[string]*websocket.Conn)

	// call grpc server
	if err := grpc.StartGRPCServer(ws, &wsMu); err != nil {
		log.Fatalf("faild to start grpc server : %v", err)
	}

	go realtime.StartWebSocketServer(ws, &wsMu)
	// block main goroutine to keep program runiing
	select {}
}
