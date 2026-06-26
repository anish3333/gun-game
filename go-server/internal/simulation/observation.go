package simulation

import (
	"github.com/anish3333/gun-game/go-server/internal/game"
)

func (s *Simulation) Observation(playerID string) []float32 {
	return game.BuildObservation(s.State, playerID)
}
