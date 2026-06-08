package engine

import (
	"math"
	"github.com/anish3333/gun-game/go-server/internal/game"
)


const CellSize = 50.0 

type SpatialHashCollision struct{}

func NewSpatialHashEngine() *SpatialHashCollision {
	return &SpatialHashCollision{}
}

// Helper to convert actual X/Y coordinates into Grid X/Y coordinates
func getCellPos(x, y float64) (int, int) {
	return int(x / CellSize), int(y / CellSize)
}

func (sh *SpatialHashCollision) ResolveCollisions(players map[string]*game.Player, bullets []*game.Bullet) []game.GameEvent {
	var events []game.GameEvent

	// 1. Build the Grid for this frame
	// We use a flat hash integer as the map key for maximum memory efficiency
	grid := make(map[int][]*game.Player)
	gridWidth := int(math.Ceil(game.ArenaW/CellSize)) + 2

	hash := func(cx, cy int) int {
		return cy*gridWidth + cx
	}

	// 2. Populate the grid: Assign players to their buckets
	for _, p := range players {
		if !p.Alive {
			continue
		}
		cx, cy := getCellPos(p.X, p.Y)
		h := hash(cx, cy)
		grid[h] = append(grid[h], p)
	}

	// 3. Resolve Bullets: Only check the bucket the bullet is in (and 8 neighbors)
	for _, b := range bullets {
		if b.Life <= 0 {
			continue
		}

		cx, cy := getCellPos(b.X, b.Y)
		hit := false

		// Scan the 3x3 grid around the bullet
		for nx := cx - 1; nx <= cx+1; nx++ {
			for ny := cy - 1; ny <= cy+1; ny++ {
				h := hash(nx, ny)
				cellPlayers := grid[h] // $O(1)$ lookup!

				for _, p := range cellPlayers {
					if !p.Alive || b.OwnerID == p.ID {
						continue
					}

					dx, dy := b.X-p.X, b.Y-p.Y
					
					// We only do this heavy math if they are in the same bucket!
					if math.Sqrt(dx*dx+dy*dy) < 22+b.Radius {
						dmg := 10
						if p.WeaponType == "pistol" { dmg = 12 }
						if p.WeaponType == "shotgun" { dmg = 18 }
						if p.WeaponType == "sniper" { dmg = 40 }
						if p.WeaponType == "smg" { dmg = 7 }

						p.HP -= dmg
						p.VX += b.VX * 0.55
						p.VY += b.VY * 0.55
						b.Life = -1
						hit = true

						if p.HP <= 0 {
							p.Alive = false
							p.HP = 0
							events = append(events, game.GameEvent{Type: "death", PlayerID: p.ID, KillerID: b.OwnerID})

							if killer, ok := players[b.OwnerID]; ok {
								killer.Score++
								if killer.Score >= game.WinLimit {
									events = append(events, game.GameEvent{Type: "match_over", PlayerID: killer.ID})
								}
							}
						} else {
							events = append(events, game.GameEvent{Type: "hit", PlayerID: p.ID, Damage: dmg, HP: p.HP})
						}
						break // Bullet is destroyed, stop checking other players
					}
				}
				if hit { break }
			}
			if hit { break }
		}
	}

	return events
}