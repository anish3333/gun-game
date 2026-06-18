package network

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/anish3333/gun-game/go-server/internal/game"
)

const (
	maxPlayers        = 2
	reconnectGrace    = 10 * time.Second
	reconnectTickRate = time.Second
)

type Room struct {
	ID            string
	HostID        string
	Phase         RoomPhase
	Config        RoomConfig
	Players       map[string]*LobbyPlayer
	Clients       map[*Client]string
	Spectators    map[*Client]struct{}
	clientByID    map[string]*Client
	spectatorByID map[string]*Client
	RematchVotes  map[string]bool
	Manager       *Manager
	State         *game.GameState

	Broadcast    chan IncomingMessage
	Register     chan *Client
	Unregister   chan *Client
	Respawn      chan string
	StartMatch   chan string
	UpdateConfig chan RoomConfig
	Rematch      chan string
	Leave        chan string
	quit         chan struct{}

	reconnectTimers map[string]*time.Timer
	mu              sync.Mutex

	playerBuf   []game.Player
	bulletBuf   []game.Bullet
	snapshotBuf SnapshotMessage
}

func NewRoom(id string, hostID string, config RoomConfig, m *Manager) *Room {
	return &Room{
		ID:              id,
		HostID:          hostID,
		Phase:           PhaseLobby,
		Config:          config,
		Players:         make(map[string]*LobbyPlayer),
		Clients:         make(map[*Client]string),
		Spectators:      make(map[*Client]struct{}),
		clientByID:      make(map[string]*Client),
		spectatorByID:   make(map[string]*Client),
		RematchVotes:    make(map[string]bool),
		Manager:         m,
		State:           game.NewGameState(m.PhysicsEngine),
		Broadcast:       make(chan IncomingMessage, 64),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client, 4),
		Respawn:         make(chan string, 4),
		StartMatch:      make(chan string, 2),
		UpdateConfig:    make(chan RoomConfig, 4),
		Rematch:         make(chan string, 4),
		Leave:           make(chan string, 4),
		quit:            make(chan struct{}),
		reconnectTimers: make(map[string]*time.Timer),
		playerBuf:       make([]game.Player, 0, maxPlayers),
		bulletBuf:       make([]game.Bullet, 0, 64),
		snapshotBuf: SnapshotMessage{
			Type:    "snapshot",
			Players: make([]game.Player, 0, maxPlayers),
			Bullets: make([]game.Bullet, 0, 64),
		},
	}
}

func (r *Room) Run() {
	ticker := time.NewTicker(time.Second / 30)
	reconnectTicker := time.NewTicker(reconnectTickRate)
	defer ticker.Stop()
	defer reconnectTicker.Stop()

	for {
		select {
		case <-r.quit:
			return

		case client := <-r.Register:
			r.handleRegister(client)

		case client := <-r.Unregister:
			r.handleUnregister(client)

		case hostID := <-r.StartMatch:
			r.handleStartMatch(hostID)

		case config := <-r.UpdateConfig:
			r.handleUpdateConfig(config)

		case playerID := <-r.Rematch:
			r.handleRematchVote(playerID)

		case playerID := <-r.Leave:
			r.handleLeave(playerID)

		case playerID := <-r.Respawn:
			r.handleRespawn(playerID)

		case message := <-r.Broadcast:
			if message.Type == "input" && r.Phase == PhasePlaying {
				if _, ok := r.State.Players[message.PlayerID]; !ok {
					continue
				}
				r.State.Inputs[message.PlayerID] = game.InputState{
					Angle: message.Angle,
					Shoot: message.Shoot,
				}
			}

		case <-reconnectTicker.C:
			if r.Phase == PhasePlaying || r.Phase == PhaseFinished {
				r.broadcastRoomState()
			}

		case <-ticker.C:
			if r.Phase != PhasePlaying {
				continue
			}

			start := time.Now()
			events := r.State.Tick()

			if r.Manager.Telemetry != nil {
				r.Manager.Telemetry.RecordTick(time.Since(start))
			}

			stillPlaying := true
			for _, ev := range events {
				switch ev.Type {
				case "match_over":
					r.handleMatchOver(ev.PlayerID)
					stillPlaying = false
				case "death":
					r.sendToAll(r.Manager.EncodeFrame(OutgoingMessage{Type: "player_died", PlayerID: ev.PlayerID, KillerID: ev.KillerID}))
					time.AfterFunc(2*time.Second, func() { r.Respawn <- ev.PlayerID })
				case "hit":
					r.sendToAll(r.Manager.EncodeFrame(OutgoingMessage{Type: "hit", PlayerID: ev.PlayerID, Damage: ev.Damage, HP: ev.HP}))
				}
			}

			if stillPlaying {
				r.sendSnapshot()
			}
		}
	}
}

func (r *Room) handleRegister(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.Players[client.ID]; ok && !existing.Connected {
		r.attachClient(client)
		r.cancelReconnectTimerLocked(client.ID)
		r.broadcastRoomStateLocked()
		if r.Phase == PhasePlaying {
			r.sendGameSyncTo(client)
			log.Printf("[%s] %s reconnected", r.ID, client.ID)
		}
		return
	}

	if len(r.Players) >= maxPlayers {
		return
	}

	r.attachClient(client)
	r.Players[client.ID] = &LobbyPlayer{
		ID:        client.ID,
		Name:      client.DisplayName,
		Weapon:    client.Weapon,
		IsHost:    client.ID == r.HostID,
		Connected: true,
	}

	log.Printf("[%s] %s joined (%s)", r.ID, client.ID, client.Weapon)
	r.broadcastRoomStateLocked()
}

func (r *Room) RegisterSpectator(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Spectators[client] = struct{}{}
	r.spectatorByID[client.ID] = client
	client.Room = r

	joinedFrame := r.Manager.EncodeFrame(OutgoingMessage{
		Type:      "room_spectated",
		Code:      r.ID,
		PlayerID:  client.ID,
		InviteURL: r.Manager.InviteURL(r.ID),
	})
	select {
	case client.Send <- joinedFrame:
	default:
	}

	r.broadcastRoomStateLocked()
	if r.Phase == PhasePlaying {
		r.sendGameSyncTo(client)
	}
	log.Printf("[%s] %s spectating", r.ID, client.ID)
}

func (r *Room) attachClient(client *Client) {
	r.Clients[client] = client.Weapon
	r.clientByID[client.ID] = client
	client.Room = r

	if lp, ok := r.Players[client.ID]; ok {
		lp.Connected = true
		lp.Weapon = client.Weapon
		lp.ReconnectSeconds = 0
		lp.ReconnectDeadline = time.Time{}
	}
}

func (r *Room) handleUnregister(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	playerID := client.ID
	wasHost := playerID == r.HostID
	delete(r.Clients, client)
	delete(r.clientByID, playerID)
	if _, ok := r.Spectators[client]; ok {
		delete(r.Spectators, client)
		delete(r.spectatorByID, playerID)
		close(client.Send)
		r.broadcastRoomStateLocked()
		return
	}

	if lp, ok := r.Players[playerID]; ok {
		lp.Connected = false
	}

	if r.Phase == PhasePlaying {
		if _, ok := r.Players[playerID]; ok {
			r.startReconnectTimerLocked(playerID)
			r.broadcastRoomStateLocked()
		}
		close(client.Send)
		return
	}

	r.cancelReconnectTimerLocked(playerID)
	r.removePlayerLocked(playerID)
	close(client.Send)

	if r.notifyAfterPlayerGoneLocked(playerID, wasHost) {
		return
	}
}

func (r *Room) handleLeave(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	wasHost := playerID == r.HostID

	if client, ok := r.spectatorByID[playerID]; ok {
		delete(r.Spectators, client)
		delete(r.spectatorByID, playerID)
		client.Room = nil
		leftMsg := r.Manager.EncodeFrame(OutgoingMessage{Type: "room_left"})
		select {
		case client.Send <- leftMsg:
		default:
		}
		r.broadcastRoomStateLocked()
		return
	}

	if client, ok := r.clientByID[playerID]; ok {
		delete(r.Clients, client)
		delete(r.clientByID, playerID)
		client.Room = nil
		leftMsg := r.Manager.EncodeFrame(OutgoingMessage{Type: "room_left"})
		select {
		case client.Send <- leftMsg:
		default:
		}
	}

	r.cancelReconnectTimerLocked(playerID)
	r.removePlayerLocked(playerID)

	r.notifyAfterPlayerGoneLocked(playerID, wasHost)
}

// notifyAfterPlayerGoneLocked tells remaining clients what happened after someone
// left or disconnected. Returns true if the room was closed.
func (r *Room) notifyAfterPlayerGoneLocked(departedID string, wasHost bool) bool {
	if wasHost || len(r.Players) == 0 {
		closedMsg := r.Manager.EncodeFrame(OutgoingMessage{
			Type:    "room_closed",
			Message: "Host left the room.",
			Reason:  "host_left",
		})
		for c := range r.Clients {
			c.Room = nil
		}
		for c := range r.Spectators {
			c.Room = nil
		}
		r.sendToAllLocked(closedMsg)
		r.closeRoomLocked()
		return true
	}

	if r.Phase == PhasePlaying {
		r.handleMatchOverLocked(r.otherPlayerID(departedID))
		return false
	}

	leftMsg := r.Manager.EncodeFrame(OutgoingMessage{
		Type:     "player_left",
		PlayerID: departedID,
	})
	r.sendToAllLocked(leftMsg)
	r.broadcastRoomStateLocked()
	return false
}

func (r *Room) handleStartMatch(hostID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Phase != PhaseLobby || hostID != r.HostID {
		return
	}
	if r.connectedCountLocked() < 2 {
		return
	}

	r.Phase = PhasePlaying
	r.State = game.NewGameState(r.Manager.PhysicsEngine)
	r.State.WinLimit = r.Config.ScoreLimit
	r.RematchVotes = make(map[string]bool)

	var pArr []game.Player
	i := 0
	for id, lp := range r.Players {
		if !lp.Connected {
			continue
		}
		spawnX := game.ArenaX + 140.0
		if i == 1 {
			spawnX = game.ArenaX + game.ArenaW - 140.0
		}

		weapon := game.Weapons[lp.Weapon]
		p := &game.Player{
			ID: id, WeaponType: lp.Weapon, Label: weapon.Label,
			X: spawnX, Y: game.ArenaY + (game.ArenaH / 2),
			VX: (rand.Float64() - 0.5) * 1.5,
			VY: (rand.Float64() - 0.5) * 1.5,
			HP: 100, Alive: true,
			FireTimer: int(float64(weapon.FireRate) * 0.5),
		}
		r.State.Players[id] = p
		pArr = append(pArr, *p)
		i++
	}

	startMsg := r.Manager.EncodeFrame(OutgoingMessage{Type: "match_start", Players: pArr})
	r.sendToAllLocked(startMsg)
	r.sendSnapshotLocked()
	log.Printf("[%s] Match started", r.ID)
}

func (r *Room) handleUpdateConfig(config RoomConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Phase != PhaseLobby {
		return
	}

	r.Config = config
	updateMsg := r.Manager.EncodeFrame(RoomConfigUpdateMessage{Type: "room_config_update", Config: r.Config})
	r.sendToAllLocked(updateMsg)
	r.broadcastRoomStateLocked()
}

func (r *Room) handleRematchVote(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Phase != PhaseFinished {
		return
	}

	r.RematchVotes[playerID] = true
	if lp, ok := r.Players[playerID]; ok {
		lp.RematchVote = true
	}

	allVoted := true
	for id, lp := range r.Players {
		if !lp.Connected {
			continue
		}
		if !r.RematchVotes[id] {
			allVoted = false
			break
		}
	}

	r.broadcastRoomStateLocked()

	if !allVoted || r.connectedCountLocked() < 2 {
		return
	}

	r.Phase = PhaseLobby
	r.State = game.NewGameState(r.Manager.PhysicsEngine)
	r.RematchVotes = make(map[string]bool)
	for _, lp := range r.Players {
		lp.RematchVote = false
	}

	r.broadcastRoomStateLocked()
	log.Printf("[%s] Rematch — back to lobby", r.ID)
}

func (r *Room) handleMatchOver(winnerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handleMatchOverLocked(winnerID)
}

func (r *Room) handleMatchOverLocked(winnerID string) {
	if r.Phase != PhasePlaying {
		return
	}

	r.Phase = PhaseFinished
	stats := r.buildMatchStatsLocked()

	resultsMsg := r.Manager.EncodeFrame(MatchResultsMessage{
		Type:     "match_results",
		WinnerID: winnerID,
		Stats:    stats,
	})
	r.sendToAllLocked(resultsMsg)

	var loserID string
	for id := range r.State.Players {
		if id != winnerID {
			loserID = id
		}
	}

	go func(winner, loser string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := r.Manager.DB.RecordMatchResult(ctx, winner, loser); err != nil {
			log.Printf("[%s] Failed to save match results: %v", r.ID, err)
		}
	}(winnerID, loserID)

	r.broadcastRoomStateLocked()
	log.Printf("[%s] Match finished. Winner: %s", r.ID, winnerID)
}

func (r *Room) buildMatchStatsLocked() map[string]PlayerMatchStats {
	stats := make(map[string]PlayerMatchStats)
	for id, p := range r.State.Players {
		deaths := 0
		for oid, op := range r.State.Players {
			if oid != id {
				deaths = op.Score
			}
		}
		stats[id] = PlayerMatchStats{Kills: p.Score, Deaths: deaths}
	}
	return stats
}

func (r *Room) handleRespawn(playerID string) {
	if p, ok := r.State.Players[playerID]; ok {
		spawnX := game.ArenaX + 140.0
		p.RespawnPlayer(spawnX, game.ArenaY+(game.ArenaH/2))

		for i := len(r.State.Bullets) - 1; i >= 0; i-- {
			if r.State.Bullets[i].OwnerID == playerID {
				r.State.Bullets = append(r.State.Bullets[:i], r.State.Bullets[i+1:]...)
			}
		}
		r.sendToAll(r.Manager.EncodeFrame(OutgoingMessage{Type: "player_respawned", PlayerID: playerID}))
	}
}

func (r *Room) startReconnectTimerLocked(playerID string) {
	r.cancelReconnectTimerLocked(playerID)

	deadline := time.Now().Add(reconnectGrace)
	if lp, ok := r.Players[playerID]; ok {
		lp.ReconnectDeadline = deadline
		lp.ReconnectSeconds = int(reconnectGrace.Seconds())
	}

	r.reconnectTimers[playerID] = time.AfterFunc(reconnectGrace, func() {
		r.ReconnectExpired(playerID)
	})
}

func (r *Room) ReconnectExpired(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lp, ok := r.Players[playerID]
	if !ok || lp.Connected {
		return
	}

	log.Printf("[%s] %s reconnect expired", r.ID, playerID)
	r.removePlayerLocked(playerID)
	delete(r.State.Players, playerID)

	if r.Phase == PhasePlaying {
		winnerID := r.otherPlayerID(playerID)
		if winnerID != "" {
			r.handleMatchOverLocked(winnerID)
		}
	} else {
		r.broadcastRoomStateLocked()
	}
}

func (r *Room) cancelReconnectTimerLocked(playerID string) {
	if t, ok := r.reconnectTimers[playerID]; ok {
		t.Stop()
		delete(r.reconnectTimers, playerID)
	}
	if lp, ok := r.Players[playerID]; ok {
		lp.ReconnectSeconds = 0
		lp.ReconnectDeadline = time.Time{}
	}
}

func (r *Room) closeRoomLocked() {
	r.Manager.DeleteRoom(r.ID)
	log.Printf("[%s] Room closed", r.ID)
	close(r.quit)
}

func (r *Room) removePlayerLocked(playerID string) {
	delete(r.Players, playerID)
	delete(r.RematchVotes, playerID)
	delete(r.State.Players, playerID)
	r.cancelReconnectTimerLocked(playerID)
}

func (r *Room) otherPlayerID(excludeID string) string {
	for id := range r.Players {
		if id != excludeID {
			return id
		}
	}
	return ""
}

func (r *Room) connectedCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.connectedCountLocked()
}

func (r *Room) connectedCountLocked() int {
	n := 0
	for _, lp := range r.Players {
		if lp.Connected {
			n++
		}
	}
	return n
}

func (r *Room) canJoin() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Phase != PhaseLobby {
		return false
	}
	return len(r.Players) < maxPlayers
}

func (r *Room) PlayerStatus(id string) (exists bool, connected bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	lp, ok := r.Players[id]
	if !ok {
		return false, false
	}
	return true, lp.Connected
}

func (r *Room) BuildRoomState() RoomStateMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buildRoomStateLocked()
}

func (r *Room) buildRoomStateLocked() RoomStateMessage {
	players := make([]LobbyPlayer, 0, len(r.Players))
	for _, lp := range r.Players {
		p := *lp
		if !p.Connected && !p.ReconnectDeadline.IsZero() {
			p.ReconnectSeconds = max(0, int(time.Until(p.ReconnectDeadline).Seconds()))
		}
		players = append(players, p)
	}

	return RoomStateMessage{
		Type:           "room_state",
		Code:           r.ID,
		Phase:          r.Phase.String(),
		HostID:         r.HostID,
		Players:        players,
		Config:         r.Config,
		InviteURL:      r.Manager.InviteURL(r.ID),
		SpectatorCount: len(r.Spectators),
	}
}

func (r *Room) broadcastRoomState() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.broadcastRoomStateLocked()
}

func (r *Room) broadcastRoomStateLocked() {
	r.sendToAllLocked(r.Manager.EncodeFrame(r.buildRoomStateLocked()))
}

func (r *Room) buildSnapshot() *SnapshotMessage {
	r.playerBuf = r.playerBuf[:0]
	r.bulletBuf = r.bulletBuf[:0]

	for _, p := range r.State.Players {
		r.playerBuf = append(r.playerBuf, *p)
	}
	for _, b := range r.State.Bullets {
		r.bulletBuf = append(r.bulletBuf, *b)
	}

	r.snapshotBuf.Players = r.playerBuf
	r.snapshotBuf.Bullets = r.bulletBuf
	return &r.snapshotBuf
}

func (r *Room) sendSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sendSnapshotLocked()
}

func (r *Room) sendSnapshotLocked() {
	r.sendToAllLocked(r.Manager.EncodeFrame(*r.buildSnapshot()))
}

func (r *Room) sendToAll(frame Frame) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sendToAllLocked(frame)
}

func (r *Room) sendToAllLocked(frame Frame) {
	for client := range r.Clients {
		select {
		case client.Send <- frame:
		default:
			close(client.Send)
			delete(r.Clients, client)
			delete(r.clientByID, client.ID)
		}
	}
	for client := range r.Spectators {
		select {
		case client.Send <- frame:
		default:
			close(client.Send)
			delete(r.Spectators, client)
			delete(r.spectatorByID, client.ID)
		}
	}
}

func (r *Room) sendGameSyncTo(client *Client) {
	var pArr []game.Player
	for _, p := range r.State.Players {
		pArr = append(pArr, *p)
	}
	startFrame := r.Manager.EncodeFrame(OutgoingMessage{Type: "match_start", Players: pArr})
	select {
	case client.Send <- startFrame:
	default:
	}

	snapFrame := r.Manager.EncodeFrame(*r.buildSnapshot())
	select {
	case client.Send <- snapFrame:
	default:
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
