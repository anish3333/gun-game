package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/anish3333/gun-game/go-server/internal/engine"
	"github.com/anish3333/gun-game/go-server/internal/game"
	"github.com/anish3333/gun-game/go-server/internal/simulation"
)

type StepRequest struct {
	Type  string  `json:"type"`
	Angle float64 `json:"angle"`
	Shoot bool    `json:"shoot"`
}

type StepResponse struct {
	Observation []float32 `json:"observation"`
	Reward      float32   `json:"reward"`
	Done        bool      `json:"done"`
}

func main() {

	sim := simulation.New(engine.NewBruteForceEngine())
	enemy := simulation.RandomEnemy{}

	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	sim.Reset()

	for scanner.Scan() {

		var req StepRequest

		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			log.Printf("bad request: %v", err)
			continue
		}

		if req.Type == "reset" {
			sim.Reset()

			encoder.Encode(StepResponse{
				Observation: sim.Observation(sim.Player1ID),
				Reward:      0,
				Done:        false,
			})

			continue
		}

		action := game.InputState{
			Angle: &req.Angle,
			Shoot: req.Shoot,
		}

		enemyAction := enemy.Act(sim)

		events := sim.Step(action, enemyAction)

		resp := StepResponse{
			Observation: sim.Observation(sim.Player1ID),
			Reward:      sim.Reward(sim.Player1ID, events),
			Done:        sim.IsDone(),
		}

		if err := encoder.Encode(resp); err != nil {
			log.Printf("encode error: %v", err)
			break
		}

		if sim.IsDone() {
			sim.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("scanner error: %v", err)
	}
}
