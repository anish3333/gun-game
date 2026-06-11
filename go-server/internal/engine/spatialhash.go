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

func getCellPos(x, y float64) (int, int) {
	return int(x / CellSize), int(y / CellSize)
}

func (sh *SpatialHashCollision) DetectCollisions(
	players map[string]*game.Player,
	bullets []*game.Bullet,
) []game.Collision {

	var collisions []game.Collision

	grid := make(map[int][]*game.Player)
	gridWidth := int(math.Ceil(game.ArenaW/CellSize)) + 2

	hash := func(cx, cy int) int {
		return cy*gridWidth + cx
	}

	for _, p := range players {
		if !p.Alive {
			continue
		}

		cx, cy := getCellPos(p.X, p.Y)
		grid[hash(cx, cy)] = append(grid[hash(cx, cy)], p)
	}

	for _, b := range bullets {

		if b.Life <= 0 {
			continue
		}

		cx, cy := getCellPos(b.X, b.Y)

		hit := false

		for nx := cx - 1; nx <= cx+1; nx++ {

			for ny := cy - 1; ny <= cy+1; ny++ {

				cellPlayers := grid[hash(nx, ny)]

				for _, p := range cellPlayers {

					if !p.Alive || p.ID == b.OwnerID {
						continue
					}

					dx, dy := b.X-p.X, b.Y-p.Y

					if math.Sqrt(dx*dx+dy*dy) < 22+b.Radius {

						collisions = append(collisions, game.Collision{
							Bullet: b,
							Target: p,
						})

						hit = true
						break
					}
				}

				if hit {
					break
				}
			}

			if hit {
				break
			}
		}
	}

	return collisions
}