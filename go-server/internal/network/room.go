package network

// Room maintains the set of active clients and broadcasts messages to them.
type Room struct {
	ID string

	// Registered clients. The boolean is just a placeholder, we use a map for O(1) lookups.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

// NewRoom acts like a constructor
func NewRoom(id string) *Room {
	return &Room{
		ID:         id,
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

// Run is the core loop. It must be started as a Goroutine!
func (r *Room) Run() {
	for {
		// The select block waits until ONE of these channels receives data.
		// This guarantees that map additions/deletions and broadcasts are 100% thread-safe.
		select {
		case client := <-r.Register:
			r.Clients[client] = true
			
		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send) // Close the client's write channel
			}
			
		case message := <-r.Broadcast:
			// A client sent a message! Broadcast it to everyone in the room.
			for client := range r.Clients {
				// We use a non-blocking send here. If a client's buffer is full, we drop them.
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}
		}
	}
}