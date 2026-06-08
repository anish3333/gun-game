package network

import (
	"sync"
	"github.com/anish3333/gun-game/go-server/internal/db"
	"github.com/anish3333/gun-game/go-server/internal/game"
	"github.com/anish3333/gun-game/go-server/internal/telemetry"
	"crypto/rand"
	"math/big"
)

type Manager struct {
	rooms         map[string]*Room
	mu            sync.RWMutex
	DB            *db.Database
	PhysicsEngine game.CollisionEngine
	Telemetry     *telemetry.Tracker    // <-- Add this
	EngineName    string                // <-- Add this so we know what is running
}

func NewManager(database *db.Database, defaultEngine game.CollisionEngine, engineName string) *Manager {
	return &Manager{
		rooms:         make(map[string]*Room),
		DB:            database,
		PhysicsEngine: defaultEngine,
		EngineName:    engineName,
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

func (m *Manager) CreateRoom() *Room {
	code := m.generateCode()
	room := NewRoom(code, m)
	
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

func (m *Manager) RoomList() []RoomSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []RoomSummary
	for _, room := range m.rooms {
		if room.Phase == "waiting" && len(room.Clients) < 2 {
			list = append(list, RoomSummary{Code: room.ID, Players: len(room.Clients)})
		}
	}
	return list
}

func (m *Manager) generateCode() string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := ""
	for i := 0; i < 4; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code += string(chars[n.Int64()])
	}
	return code
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
		count += len(r.Clients)
	}
	return count
}

func (m *Manager) GetEngineName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.EngineName
}