package simulation

import (
	"math/rand"

	"github.com/anish3333/gun-game/go-server/internal/game"
)

type Simulation struct {
	State *game.GameState

	Player1ID string
	Player2ID string

	Done bool
}

func New(physics game.CollisionEngine) *Simulation {
	s := &Simulation{
		State: game.NewGameState(physics),
	}

	s.Reset()

	return s
}

func (s *Simulation) Reset() {
	s.Done = false

	s.State.Players = make(map[string]*game.Player)
	s.State.Bullets = nil
	s.State.Inputs = make(map[string]game.InputState)

	s.Player1ID = "p1"
	s.Player2ID = "p2"

	s.State.Players[s.Player1ID] = &game.Player{
		ID:         s.Player1ID,
		WeaponType: "pistol",
		Label:      game.Weapons["pistol"].Label,
		X:          game.ArenaX + 140,
		Y:          game.ArenaY + game.ArenaH/2,
		VX:         (rand.Float64() - 0.5) * 1.5,
		VY:         (rand.Float64() - 0.5) * 1.5,
		HP:         100,
		Alive:      true,
	}

	s.State.Players[s.Player2ID] = &game.Player{
		ID:         s.Player2ID,
		WeaponType: "pistol",
		Label:      game.Weapons["pistol"].Label,
		X:          game.ArenaX + game.ArenaW - 140,
		Y:          game.ArenaY + game.ArenaH/2,
		VX:         (rand.Float64() - 0.5) * 1.5,
		VY:         (rand.Float64() - 0.5) * 1.5,
		HP:         100,
		Alive:      true,
	}
}

func (s *Simulation) Step(
	input1 game.InputState,
	input2 game.InputState,
) []game.GameEvent {

	s.State.Inputs[s.Player1ID] = input1
	s.State.Inputs[s.Player2ID] = input2

	events := s.State.Tick()

	for _, ev := range events {
		if ev.Type == "death" {
			s.Done = true
		}
	}

	return events
}

func (s *Simulation) IsDone() bool {
	return s.Done
}
