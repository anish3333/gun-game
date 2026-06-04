package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/anish3333/gun-game/go-server/internal/auth"
	"github.com/anish3333/gun-game/go-server/internal/db"
	"github.com/anish3333/gun-game/go-server/internal/network"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var gameManager = network.NewManager()

// 1. Global DB reference
var database *db.Database

func main() {
	// 1. Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file. Falling back to system environment variables.")
	}
	// 2. Initialize Postgres (Replace with your actual Postgres URL)
	dsn := os.Getenv("DATABASE_URL")
	database = db.InitDB(dsn)
	
	// Gracefully close the pool when the server shuts down
	defer database.Pool.Close()

	port := os.Getenv("PORT")
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/init-guest", InitGuestHandler)

	log.Printf("✦ Recoil Arena GO server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitGuestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	token, playerID, displayName, err := auth.GenerateGuestToken()
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// 3. Save the new identity permanently in PostgreSQL!
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	err = database.CreatePlayer(ctx, playerID, displayName)
	if err != nil {
		log.Printf("DB Error creating player: %v", err)
		// We still return the token so the player can play even if the DB blipped
	}

	response := struct {
		Token       string `json:"token"`
		PlayerID    string `json:"player_id"`
		DisplayName string `json:"display_name"`
	}{
		Token:       token,
		PlayerID:    playerID,
		DisplayName: displayName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateGuestToken(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil { return }

	// 4. Fire-and-forget DB update for last login
	go database.UpdateLastLogin(context.Background(), claims.PlayerID)

	client := &network.Client{
		ID:      claims.PlayerID,
		Manager: gameManager,
		Conn:    ws,
		Send:    make(chan []byte, 256),
	}

	log.Printf("Verified Player connected: %s (%s)", client.ID, claims.DisplayName)

	go client.WritePump()
	client.Send <- network.BuildHello(gameManager)
	go client.ReadPump()
}