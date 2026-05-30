package main

import (
	"log"
	"net/http"
    "fmt"
    "time"

	"github.com/gorilla/websocket"
	
	"github.com/anish3333/gun-game/go-server/internal/network"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Create a global test room for now
var lobby = network.NewRoom("lobby-1")

func main() {
	// Start the room loop in its own concurrent Goroutine
	go lobby.Run()

	port := ":3000"
	http.HandleFunc("/ws", handleConnections)

	log.Printf("✦ Recoil Arena GO server starting on ws://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Fatal error upgrading connection: %v", err)
		return
	}

	// Create a new client instance
	client := &network.Client{
		ID:   fmt.Sprintf("p%d", time.Now().UnixNano()), // Temp ID generation
		Room: lobby,
		Conn: ws,
		Send: make(chan []byte, 256), // Buffered channel size
	}

	// Register the client with the room
	client.Room.Register <- client

	// Start the pumps in concurrent Goroutines
	go client.WritePump()
	go client.ReadPump()
	
	log.Printf("Client %s connected to Lobby!", client.ID)
}