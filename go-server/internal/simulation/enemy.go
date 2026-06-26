package simulation

import "github.com/anish3333/gun-game/go-server/internal/game"

type Enemy interface {
	Act(sim *Simulation) game.InputState
}
