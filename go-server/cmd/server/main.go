package main

import (
	"log"

	"github.com/anish3333/gun-game/go-server/internal/config"
	"github.com/anish3333/gun-game/go-server/internal/db"
	"github.com/anish3333/gun-game/go-server/internal/engine"
	"github.com/anish3333/gun-game/go-server/internal/network"
	"github.com/anish3333/gun-game/go-server/internal/server"
	"github.com/anish3333/gun-game/go-server/internal/telemetry"
)

func main() {

	cfg := config.Load()

	database := db.InitDB(cfg.DatabaseURL)
	defer database.Pool.Close()

	tracker := telemetry.NewTracker()

	physics := engine.NewBruteForceEngine()

	manager := network.NewManager(
		database,
		physics,
		"BruteForce O(N²)",
	)

	manager.Telemetry = tracker

	tracker.GetRoomCount = manager.GetRoomCount
	tracker.GetClientCount = manager.GetTotalClients
	tracker.GetStrategy = manager.GetEngineName

	go tracker.Start()

	srv := server.New(
		cfg,
		database,
		manager,
		tracker,
	)

	log.Printf("Server listening on %s", cfg.Port)

	log.Fatal(srv.Run())
}
