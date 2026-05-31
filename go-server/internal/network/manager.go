package network

import (
	"crypto/rand"
	"math/big"
	"sync"
)

type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
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