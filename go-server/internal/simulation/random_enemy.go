package simulation

import (
	"math"
	"math/rand"

	"github.com/anish3333/gun-game/go-server/internal/game"
)

type RandomEnemy struct{}

func (e RandomEnemy) Act(sim *Simulation) game.InputState {

	// player := sim.State.Players[sim.Player2ID]

	angle := rand.Float64() * 2 * math.Pi

	return game.InputState{
		Angle: &angle,
		Shoot: rand.Float64() < 0.1,
	}
}
