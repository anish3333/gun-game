package game


type Collision struct {
	Bullet *Bullet
	Target *Player
}

type CollisionEngine interface {
	DetectCollisions(
		players map[string]*Player,
		bullets []*Bullet,
	) []Collision
}