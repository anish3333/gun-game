package telemetry

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ServerMetrics is the exact JSON payload the Admin Dashboard will graph
type ServerMetrics struct {
	Goroutines  int     `json:"goroutines"`
	HeapAllocMB float64 `json:"heapAllocMb"`
	TickTimeUs  int64   `json:"tickTimeUs"` // Microseconds (1ms = 1000us)
	RoomCount   int     `json:"roomCount"`
	ClientCount int     `json:"clientCount"`
	Strategy    string  `json:"strategy"`
}

type Tracker struct {
	admins map[*websocket.Conn]bool
	mu     sync.RWMutex

	// Tick tracking (updated by hundreds of rooms 30x a second)
	tickTimes []time.Duration
	tickMu    sync.Mutex

	// Callbacks to prevent circular imports with the network package
	GetRoomCount   func() int
	GetClientCount func() int
	GetStrategy    func() string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewTracker() *Tracker {
	return &Tracker{
		admins:    make(map[*websocket.Conn]bool),
		tickTimes: make([]time.Duration, 0, 1000),
	}
}

// RecordTick is called by every active room after it runs physics
func (t *Tracker) RecordTick(duration time.Duration) {
	t.tickMu.Lock()
	t.tickTimes = append(t.tickTimes, duration)
	t.tickMu.Unlock()
}

// Start opens a 1-second background loop to calculate and broadcast metrics
func (t *Tracker) Start() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		t.broadcastMetrics()
	}
}

func (t *Tracker) broadcastMetrics() {
	t.mu.RLock()
	if len(t.admins) == 0 {
		t.mu.RUnlock()
		return // Save CPU if no admins are watching
	}
	t.mu.RUnlock()

	// 1. Calculate Average Tick Time
	t.tickMu.Lock()
	var totalUs int64
	count := len(t.tickTimes)
	for _, d := range t.tickTimes {
		totalUs += d.Microseconds()
	}
	t.tickTimes = t.tickTimes[:0] // Reset slice without reallocating memory!
	t.tickMu.Unlock()

	avgUs := int64(0)
	if count > 0 {
		avgUs = totalUs / int64(count)
	}

	// 2. Read Go System Memory (The SDE 2 Flex)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := ServerMetrics{
		Goroutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(m.Alloc) / 1024 / 1024,
		TickTimeUs:  avgUs,
		RoomCount:   0,
		ClientCount: 0,
		Strategy:    "unknown",
	}

	// Safely fetch dynamic game stats
	if t.GetRoomCount != nil { metrics.RoomCount = t.GetRoomCount() }
	if t.GetClientCount != nil { metrics.ClientCount = t.GetClientCount() }
	if t.GetStrategy != nil { metrics.Strategy = t.GetStrategy() }

	payload, _ := json.Marshal(metrics)

	// 3. Broadcast to all connected admins
	t.mu.Lock()
	for admin := range t.admins {
		if err := admin.WriteMessage(websocket.TextMessage, payload); err != nil {
			admin.Close()
			delete(t.admins, admin)
		}
	}
	t.mu.Unlock()
}

// HandleAdminWS connects the Admin Dashboard to the telemetry stream
func (t *Tracker) HandleAdminWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Admin WS Upgrade Error: %v", err)
		return
	}

	t.mu.Lock()
	t.admins[ws] = true
	t.mu.Unlock()
	log.Println("✦ Admin Dashboard Connected to Telemetry")
}