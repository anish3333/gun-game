package network

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/anish3333/gun-game/go-server/internal/game"
)

type Room struct {
	ID         string
	Phase      string
	Clients    map[*Client]string // Client -> Weapon mapping
	Manager    *Manager
	State      *game.GameState
	
	// Channels (Thread-safe pipes)
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Respawn    chan string // Safely triggers respawns from async timers
}

func NewRoom(id string, m *Manager) *Room {
	return &Room{
		ID:         id,
		Phase:      "waiting",
		Clients:    make(map[*Client]string),
		Manager:    m,
		State:      game.NewGameState(),
		Broadcast:  make(chan []byte, 64),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Respawn:    make(chan string),
	}
}

func (r *Room) Run() {
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = client.Weapon
			
			// Match Start Logic
			if len(r.Clients) == 2 && r.Phase == "waiting" {
				r.Phase = "playing"
				var pArr []game.Player
				i := 0
				
				for c, w := range r.Clients {
					spawnX := game.ArenaX + 140.0
					if i == 1 {
						spawnX = game.ArenaX + game.ArenaW - 140.0
					}

					weapon := game.Weapons[w]
					p := &game.Player{
						ID: c.ID, WeaponType: w, Label: weapon.Label,
						X: spawnX, Y: game.ArenaY + (game.ArenaH / 2),
						VX: (rand.Float64() - 0.5) * 1.5,
						VY: (rand.Float64() - 0.5) * 1.5,
						HP: 100, Alive: true,
						FireTimer: int(float64(weapon.FireRate) * 0.5),
					}
					r.State.Players[c.ID] = p
					pArr = append(pArr, *p)
					i++
				}

				startMsg, _ := json.Marshal(OutgoingMessage{Type: "match_start", Players: pArr})
				r.sendToAll(startMsg)
				r.sendToAll(r.buildSnapshot())
				log.Printf("[%s] Match Started", r.ID)
			}

		case client := <-r.Unregister:
			delete(r.Clients, client)
			delete(r.State.Players, client.ID)
			close(client.Send)
			
			if r.Phase == "playing" {
				discMsg, _ := json.Marshal(OutgoingMessage{Type: "opponent_disconnected"})
				r.sendToAll(discMsg)
			}
			// Cleanup if empty
			if len(r.Clients) == 0 {
				r.Manager.DeleteRoom(r.ID)
				log.Printf("[%s] Room closed", r.ID)
				return // Ends the goroutine
			}

		case playerID := <-r.Respawn:
			if p, ok := r.State.Players[playerID]; ok {
				// Determine spawn side based on map order (simplified)
				spawnX := game.ArenaX + 140.0
				p.RespawnPlayer(spawnX, game.ArenaY+(game.ArenaH/2))
				
				// Clear their bullets
				for i := len(r.State.Bullets) - 1; i >= 0; i-- {
					if r.State.Bullets[i].OwnerID == playerID {
						r.State.Bullets = append(r.State.Bullets[:i], r.State.Bullets[i+1:]...)
					}
				}
				resMsg, _ := json.Marshal(OutgoingMessage{Type: "player_respawned", PlayerID: playerID})
				r.sendToAll(resMsg)
			}

		case message := <-r.Broadcast:
			var msg IncomingMessage
			if err := json.Unmarshal(message, &msg); err == nil && msg.Type == "input" && r.Phase == "playing" {
				r.State.Inputs[msg.PlayerID] = game.InputState{Angle: msg.Angle, Shoot: msg.Shoot}
			}

		case <-ticker.C:
			if r.Phase != "playing" { continue }

			// 1. Calculate Math
			events := r.State.Tick()

			// 2. Handle Events
			for _, ev := range events {
				if ev.Type == "death" {
					deathMsg, _ := json.Marshal(OutgoingMessage{Type: "player_died", PlayerID: ev.PlayerID, KillerID: ev.KillerID})
					r.sendToAll(deathMsg)
					// Safe async trigger to respawn channel
					time.AfterFunc(2*time.Second, func() { r.Respawn <- ev.PlayerID })
				} else if ev.Type == "hit" {
					hitMsg, _ := json.Marshal(OutgoingMessage{Type: "hit", PlayerID: ev.PlayerID, Damage: ev.Damage, HP: ev.HP})
					r.sendToAll(hitMsg)
				}
			}

			r.sendToAll(r.buildSnapshot())
		}
	}
}

func (r *Room) buildSnapshot() []byte {
	pArr := make([]game.Player, 0, len(r.State.Players))
	for _, p := range r.State.Players {
		pArr = append(pArr, *p)
	}
	bArr := make([]game.Bullet, 0, len(r.State.Bullets))
	for _, b := range r.State.Bullets {
		bArr = append(bArr, *b)
	}
	snap, _ := json.Marshal(SnapshotMessage{Type: "snapshot", Players: pArr, Bullets: bArr})
	return snap
}

func (r *Room) sendToAll(msg []byte) {
	for client := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}