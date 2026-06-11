package server

import (
	"log"
	"net/http"

	"github.com/anish3333/gun-game/go-server/internal/engine"
)

func (s *Server) AdminSwapStrategyHandler(w http.ResponseWriter, r *http.Request) {
	// Simple CORS for the admin dashboard
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions { return }

	strategy := r.URL.Query().Get("type")
	
	if strategy == "spatial" {
		s.manager.SetCollisionEngine(engine.NewSpatialHashEngine(), "SpatialHash O(N)")
		log.Println("✦ Admin Hot-Swapped Engine -> SpatialHash")
	} else {
		s.manager.SetCollisionEngine(engine.NewBruteForceEngine(), "BruteForce O(N^2)")
		log.Println("✦ Admin Hot-Swapped Engine -> BruteForce")
	}
	
	w.WriteHeader(http.StatusOK)
}