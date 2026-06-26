package simulation

import "github.com/anish3333/gun-game/go-server/internal/game"

func (s *Simulation) Reward(playerID string, events []game.GameEvent) float32 {

	reward := float32(-0.01)

	for _, ev := range events {

		switch ev.Type {

		case "hit":
			if ev.PlayerID != playerID {
				reward += 5
			} else {
				reward -= 5
			}

		case "death":
			if ev.PlayerID == playerID {
				reward -= 100
			}
			if ev.KillerID == playerID {
				reward += 100
			}

		case "match_over":
			if ev.PlayerID == playerID {
				reward += 1000
			}
		}
	}

	return reward
}
