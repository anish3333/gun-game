package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/anish3333/gun-game/go-server/internal/auth"
)

func (s *Server) InitGuestHandler(w http.ResponseWriter, r *http.Request) {
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
	
	err = s.db.CreatePlayer(ctx, playerID, displayName)
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