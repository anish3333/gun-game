package game

import "math"

func normalizeAngle(x float64) float64 {
	for x > math.Pi {
		x -= 2 * math.Pi
	}
	for x < -math.Pi {
		x += 2 * math.Pi
	}
	return x
}

func BuildObservation(state *GameState, playerID string) []float32 {
	player := state.Players[playerID]

	var enemy *Player
	for id, p := range state.Players {
		if id != playerID {
			enemy = p
			break
		}
	}

	dx := enemy.X - player.X
	dy := enemy.Y - player.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	angleToEnemy := math.Atan2(dy, dx)
	angleDiff := normalizeAngle(angleToEnemy - player.Angle)

	return []float32{
		// self
		float32(player.X),
		float32(player.Y),
		float32(player.VX),
		float32(player.VY),
		float32(player.Angle),
		float32(player.HP),

		// enemy
		float32(enemy.X),
		float32(enemy.Y),
		float32(enemy.VX),
		float32(enemy.VY),
		float32(enemy.Angle),
		float32(enemy.HP),

		// relative
		float32(dx),
		float32(dy),
		float32(distance),
		float32(angleDiff),
	}
}
