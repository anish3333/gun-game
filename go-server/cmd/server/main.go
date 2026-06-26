package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"

	"github.com/anish3333/gun-game/go-server/internal/config"
	"github.com/anish3333/gun-game/go-server/internal/db"
	"github.com/anish3333/gun-game/go-server/internal/engine"
	"github.com/anish3333/gun-game/go-server/internal/game"
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

	// Use the project's `rl` virtualenv and predictor script.
	// When running the server from `go-server/`, these relative paths point to the
	// venv python and the `predict.py` in the `rl/` folder.
	ppoPython := "../rl/.venv/bin/python"
	ppoScript := "../rl/predict.py"
	ppoCmd := exec.Command(ppoPython, ppoScript)
	ppoCmd.Dir = "../rl"
	ppoStdin, err := ppoCmd.StdinPipe()
	if err != nil {
		log.Printf("failed to create PPO stdin pipe: %v", err)
	} else {
		ppoStdout, err := ppoCmd.StdoutPipe()
		if err != nil {
			log.Printf("failed to create PPO stdout pipe: %v", err)
		} else {
			ppoCmd.Stderr = os.Stderr
			if err := ppoCmd.Start(); err != nil {
				log.Printf("failed to start PPO process: %v", err)
			} else {
				manager.PPOController = game.NewPPOController(ppoStdin, bufio.NewReader(ppoStdout))
				log.Printf("PPO process started (%s %s)", ppoPython, ppoScript)
			}
		}
	}

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
