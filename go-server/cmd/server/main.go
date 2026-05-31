package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/anish3333/gun-game/go-server/internal/network"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var gameManager = network.NewManager()

func main() {
	port := ":3000"
	http.HandleFunc("/ws", handleConnections)

	log.Printf("✦ Recoil Arena GO server starting on ws://localhost%s", port)
	log.Printf("  Tick rate: 30/sec | Fully Concurrent")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &network.Client{
		ID:      fmt.Sprintf("p%d", time.Now().UnixNano()),
		Manager: gameManager,
		Conn:    ws,
		Send:    make(chan []byte, 256),
	}

	log.Printf("Client connected: %s", client.ID)

	go client.WritePump()
	client.Send <- network.BuildHello(gameManager)
	go client.ReadPump()
}