package network

import (
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/anish3333/gun-game/go-server/internal/codec"
	"github.com/anish3333/gun-game/go-server/internal/db"
	"github.com/anish3333/gun-game/go-server/internal/game"
	"github.com/anish3333/gun-game/go-server/internal/telemetry"
)

type Manager struct {
	rooms         map[string]*Room
	clients       map[*Client]struct{}
	mu            sync.RWMutex
	DB            *db.Database
	PhysicsEngine game.CollisionEngine
	Telemetry     *telemetry.Tracker
	EngineName    string
	BaseURL       string
	Codec         codec.Codec
	CodecName     string
	PPOController *game.PPOController
}

func NewManager(database *db.Database, defaultEngine game.CollisionEngine, engineName string, baseURL string, encoding string) *Manager {
	c := codec.New(encoding)
	return &Manager{
		rooms:         make(map[string]*Room),
		clients:       make(map[*Client]struct{}),
		DB:            database,
		PhysicsEngine: defaultEngine,
		EngineName:    engineName,
		BaseURL:       baseURL,
		Codec:         c,
		CodecName:     c.Name(),
	}
}

func (m *Manager) SetCollisionEngine(newEngine game.CollisionEngine, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PhysicsEngine = newEngine
	m.EngineName = name

	for _, r := range m.rooms {
		r.State.Physics = newEngine
	}
}

func (m *Manager) RegisterClient(c *Client) {
	m.mu.Lock()
	m.clients[c] = struct{}{}
	m.mu.Unlock()
}

func (m *Manager) UnregisterClient(c *Client) {
	m.mu.Lock()
	delete(m.clients, c)
	m.mu.Unlock()
}

func (m *Manager) SetCodec(name string) {
	newCodec := codec.New(name)

	m.mu.Lock()
	m.Codec = newCodec
	m.CodecName = newCodec.Name()
	clients := make([]*Client, 0, len(m.clients))
	for c := range m.clients {
		clients = append(clients, c)
	}
	m.mu.Unlock()

	ctrl := m.EncodeControlJSON(EncodingChangedMessage{
		Type:     "encoding_changed",
		Encoding: newCodec.Name(),
	})
	for _, c := range clients {
		select {
		case c.Send <- ctrl:
		default:
		}
	}
}

func (m *Manager) GetCodecName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.CodecName
}

func (m *Manager) CreateRoom(hostID string, config RoomConfig) *Room {
	code := m.generateCode()
	room := NewRoom(code, hostID, config, m)

	m.mu.Lock()
	m.rooms[code] = room
	m.mu.Unlock()

	go room.Run()
	return room
}

func (m *Manager) GetRoom(code string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, exists := m.rooms[code]
	return room, exists
}

func (m *Manager) DeleteRoom(code string) {
	m.mu.Lock()
	delete(m.rooms, code)
	m.mu.Unlock()
}

func (m *Manager) RoomInfo(code string) RoomInfoResponse {
	m.mu.RLock()
	room, exists := m.rooms[code]
	m.mu.RUnlock()

	if !exists {
		return RoomInfoResponse{Exists: false}
	}

	connected := room.connectedCount()
	return RoomInfoResponse{
		Exists:      true,
		Code:        room.ID,
		Phase:       room.Phase.String(),
		PlayerCount: connected,
		CanJoin:     room.canJoin(),
		CanSpectate: true,
	}
}

func (m *Manager) generateCode() string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for {
		code := ""
		for i := 0; i < 6; i++ {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
			code += string(chars[n.Int64()])
		}
		m.mu.RLock()
		_, taken := m.rooms[code]
		m.mu.RUnlock()
		if !taken {
			return code
		}
	}
}

func (m *Manager) GetRoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

func (m *Manager) GetTotalClients() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, r := range m.rooms {
		count += len(r.Clients) + len(r.Spectators)
	}
	return count
}

func (m *Manager) GetEngineName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.EngineName
}

func (m *Manager) InviteURL(code string) string {
	if m.BaseURL != "" {
		return m.BaseURL + "/play/" + code
	}
	return "/play/" + code
}
