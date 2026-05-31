package game

import "math"

const (
	ArenaX, ArenaY = 30.0, 30.0
	ArenaW, ArenaH = 620.0, 420.0
	Gravity        = 0.04
	Damping        = 0.992
	WallBounce     = 0.82
	GunRadius      = 32.0
	BulletDecay    = 0.010
)

type WeaponDef struct {
	Label        string  `json:"label"`
	Type         string  `json:"type"`
	FireRate     int     `json:"fireRate"`
	RecoilForce  float64 `json:"recoilForce"`
	BulletSpeed  float64 `json:"bulletSpeed"`
	BulletRadius float64 `json:"bulletRadius"`
	Pellets      int     `json:"pellets"`
	Spread       float64 `json:"spread"`
	Color        string  `json:"color"`
	Desc         string  `json:"desc"`
}

var Weapons = map[string]WeaponDef{
	"pistol": {
		Label: "PULSAR-9", Type: "pistol",
		FireRate: 20, RecoilForce: 3.6, BulletSpeed: 9,
		BulletRadius: 2, Pellets: 1, Spread: 0.06, Color: "#4af0c8",
		Desc: "fast fire · floaty kicks · high mobility",
	},
	"shotgun": {
		Label: "SLEDGE-X", Type: "shotgun",
		FireRate: 58, RecoilForce: 10, BulletSpeed: 7,
		BulletRadius: 3, Pellets: 4, Spread: 0.38, Color: "#f0a84a",
		Desc: "massive kick · spread shot · slow rate",
	},
	"smg": {
		Label: "WASP-7", Type: "smg",
		FireRate: 8, RecoilForce: 1.8, BulletSpeed: 11,
		BulletRadius: 1.5, Pellets: 1, Spread: 0.10, Color: "#a84af0",
		Desc: "rapid micro-kicks · constant drift · fast bullets",
	},
	"sniper": {
		Label: "LANCE-1", Type: "sniper",
		FireRate: 90, RecoilForce: 16, BulletSpeed: 18,
		BulletRadius: 2, Pellets: 1, Spread: 0.01, Color: "#f04a4a",
		Desc: "one giant kick · extreme range · high damage",
	},
}

type Player struct {
	ID          string  `json:"id"`
	WeaponType  string  `json:"weaponType"`
	Label       string  `json:"label"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	VX          float64 `json:"vx"`
	VY          float64 `json:"vy"`
	Angle       float64 `json:"angle"`
	HP          int     `json:"hp"`
	Alive       bool    `json:"alive"`
	Score       int     `json:"score"`
	MuzzleFlash float64 `json:"muzzleFlash"`
	FireTimer   int     `json:"-"`
}

type Bullet struct {
	ID      int     `json:"id"`
	OwnerID string  `json:"ownerId"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	VX      float64 `json:"vx"`
	VY      float64 `json:"vy"`
	Radius  float64 `json:"r"`
	Life    float64 `json:"-"`
}

type InputState struct {
	Angle *float64; Shoot bool
}

// GameEvent lets the engine pass hit/death notifications back to the room
type GameEvent struct {
	Type     string; PlayerID string; KillerID string; Damage int; HP int
}

type GameState struct {
	Players         map[string]*Player
	Bullets         []*Bullet
	Inputs          map[string]InputState
	BulletIDCounter int
}

func NewGameState() *GameState {
	return &GameState{
		Players: make(map[string]*Player),
		Inputs:  make(map[string]InputState),
	}
}

func (state *GameState) Tick() []GameEvent {
	var events []GameEvent

	// 1. Process Players
	for id, player := range state.Players {
		if !player.Alive { continue }

		weapon := Weapons[player.WeaponType]
		input := state.Inputs[id]
		if input.Angle != nil { player.Angle = *input.Angle }

		player.FireTimer++
		if input.Shoot && player.FireTimer >= weapon.FireRate {
			player.FireTimer = 0
			player.MuzzleFlash = 1.0

			mx := player.X + math.Cos(player.Angle)*GunRadius
			my := player.Y + math.Sin(player.Angle)*GunRadius

			for i := 0; i < weapon.Pellets; i++ {
				spread := (math.Sin(float64(state.BulletIDCounter+1))*10000 - math.Floor(math.Sin(float64(state.BulletIDCounter+1))*10000) - 0.5) * weapon.Spread
				fa := player.Angle + spread

				state.Bullets = append(state.Bullets, &Bullet{
					ID: state.BulletIDCounter, OwnerID: player.ID,
					X: mx, Y: my, VX: math.Cos(fa)*weapon.BulletSpeed, VY: math.Sin(fa)*weapon.BulletSpeed,
					Radius: weapon.BulletRadius, Life: 1.0,
				})
				state.BulletIDCounter++
			}
			player.VX -= math.Cos(player.Angle) * weapon.RecoilForce
			player.VY -= math.Sin(player.Angle) * weapon.RecoilForce
		}

		player.MuzzleFlash = math.Max(0, player.MuzzleFlash-0.10)
		player.VY += Gravity
		player.VX *= Damping
		player.VY *= Damping
		player.X += player.VX
		player.Y += player.VY

		minX, maxX := ArenaX+GunRadius, ArenaX+ArenaW-GunRadius
		minY, maxY := ArenaY+GunRadius, ArenaY+ArenaH-GunRadius
		if player.X < minX { player.X = minX; player.VX = math.Abs(player.VX) * WallBounce }
		if player.X > maxX { player.X = maxX; player.VX = -math.Abs(player.VX) * WallBounce }
		if player.Y < minY { player.Y = minY; player.VY = math.Abs(player.VY) * WallBounce }
		if player.Y > maxY { player.Y = maxY; player.VY = -math.Abs(player.VY) * WallBounce }
	}

	// 2. Process Bullets & Hit Detection (Reverse loop for safe deletion)
	for i := len(state.Bullets) - 1; i >= 0; i-- {
		b := state.Bullets[i]
		b.X += b.VX; b.Y += b.VY; b.Life -= BulletDecay

		if b.X < ArenaX { b.VX *= -0.9; b.X = ArenaX }
		if b.X > ArenaX+ArenaW { b.VX *= -0.9; b.X = ArenaX + ArenaW }
		if b.Y < ArenaY { b.VY *= -0.9; b.Y = ArenaY }
		if b.Y > ArenaY+ArenaH { b.VY *= -0.9; b.Y = ArenaY + ArenaH }

		for _, p := range state.Players {
			if !p.Alive || b.OwnerID == p.ID || b.Life <= 0 { continue }
			
			dx, dy := b.X-p.X, b.Y-p.Y
			if math.Sqrt(dx*dx+dy*dy) < 22+b.Radius {
				dmg := 10 // default
				if p.WeaponType == "pistol" { dmg = 12 } else if p.WeaponType == "shotgun" { dmg = 18 } else if p.WeaponType == "sniper" { dmg = 40 } else if p.WeaponType == "smg" { dmg = 7 }
				
				p.HP -= dmg
				p.VX += b.VX * 0.55
				p.VY += b.VY * 0.55
				b.Life = -1
				
				if p.HP <= 0 {
					p.Alive = false; p.HP = 0
					if killer, ok := state.Players[b.OwnerID]; ok { killer.Score++ }
					events = append(events, GameEvent{Type: "death", PlayerID: p.ID, KillerID: b.OwnerID})
				} else {
					events = append(events, GameEvent{Type: "hit", PlayerID: p.ID, Damage: dmg, HP: p.HP})
				}
			}
		}
		if b.Life <= 0 { state.Bullets = append(state.Bullets[:i], state.Bullets[i+1:]...) }
	}

	state.Inputs = make(map[string]InputState) // Clear inputs
	return events
}

// RespawnPlayer resets stats
func (p *Player) RespawnPlayer(x, y float64) {
	p.X = x; p.Y = y
	p.VX, p.VY = 0, 0
	p.HP = 100
	p.Alive = true
	weapon := Weapons[p.WeaponType]
	p.FireTimer = int(float64(weapon.FireRate) * 0.4)
}