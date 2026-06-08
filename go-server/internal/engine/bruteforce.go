package engine

import (
	"math"
	"github.com/anish3333/gun-game/go-server/internal/game"
)

// BruteForceCollision implements game.CollisionEngine
type BruteForceCollision struct{}

func NewBruteForceEngine() *BruteForceCollision {
	return &BruteForceCollision{}
}

// ResolveCollisions iterates every bullet against every player O(N*M)
func (bf *BruteForceCollision) ResolveCollisions(players map[string]*game.Player, bullets []*game.Bullet) []game.GameEvent {
	var events []game.GameEvent

	for _, b := range bullets {
		if b.Life <= 0 {
			continue
		}

		for _, p := range players {
			if !p.Alive || b.OwnerID == p.ID {
				continue
			}

			dx, dy := b.X-p.X, b.Y-p.Y
			// Distance check
			if math.Sqrt(dx*dx+dy*dy) < 22+b.Radius {
				// Calculate Damage based on weapon
				dmg := 10 // default
				if p.WeaponType == "pistol" {
					dmg = 12
				} else if p.WeaponType == "shotgun" {
					dmg = 18
				} else if p.WeaponType == "sniper" {
					dmg = 40
				} else if p.WeaponType == "smg" {
					dmg = 7
				}

				p.HP -= dmg
				p.VX += b.VX * 0.55
				p.VY += b.VY * 0.55
				b.Life = -1 // Mark for cleanup

				// Process Death & Win Conditions
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
			}
		}
	}

	return events
}