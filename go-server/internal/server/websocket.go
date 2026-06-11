package server

import (
	"context"
	"log"
	"net/http"

	"github.com/anish3333/gun-game/go-server/internal/auth"
	"github.com/anish3333/gun-game/go-server/internal/network"
)
func (s *Server) HandleConnections(w http.ResponseWriter, r *http.Request) {
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

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil { return }

	// 4. Fire-and-forget DB update for last login
	go s.db.UpdateLastLogin(context.Background(), claims.PlayerID)

	client := &network.Client{
		ID:      claims.PlayerID,
		Manager: s.manager,
		Conn:    ws,
		Send:    make(chan []byte, 256),
	}

	log.Printf("Verified Player connected: %s (%s)", client.ID, claims.DisplayName)

	go client.WritePump()
	client.Send <- network.BuildHello(s.manager)
	go client.ReadPump()
}