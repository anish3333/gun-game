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

	encoding := cfg.DefaultEncoding

	manager := network.NewManager(
		database,
		physics,
		"BruteForce O(N²)",
		cfg.BaseURL,
		encoding,
	)

	manager.Telemetry = tracker

	tracker.GetRoomCount = manager.GetRoomCount
	tracker.GetClientCount = manager.GetTotalClients
	tracker.GetStrategy = manager.GetEngineName
	tracker.GetEncoding = manager.GetCodecName

	go tracker.Start()

	srv := server.New(
		cfg,
		database,
		manager,
		tracker,
	)

	log.Printf("Server listening on %s (encoding: %s)", cfg.Port, encoding)

	log.Fatal(srv.Run())
}
