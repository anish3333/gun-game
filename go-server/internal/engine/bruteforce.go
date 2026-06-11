package engine

import (
	"math"

	"github.com/anish3333/gun-game/go-server/internal/game"
)

type BruteForceCollision struct{}

func NewBruteForceEngine() *BruteForceCollision {
	return &BruteForceCollision{}
}

func (bf *BruteForceCollision) DetectCollisions(
	players map[string]*game.Player,bullets []*game.Bullet,	
) []game.Collision {

	var collisions []game.Collision
	for _, b := range bullets {
		if b.Life <= 0 { continue }
		for _, p := range players {
			if !p.Alive || b.OwnerID == p.ID { continue }
			dx, dy := b.X-p.X, b.Y-p.Y
			if math.Sqrt(dx*dx+dy*dy) < 22+b.Radius {
				collisions = append(collisions, game.Collision{ Bullet: b, Target: p })
				break
			}
		}
	}

	return collisions
}