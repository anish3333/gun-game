package network

import (
	"log"
	// "time"
	"github.com/gorilla/websocket"
)

// Client acts as the middleman between the websocket connection and the game room.
type Client struct {
	ID   string
	Room *Room
	Conn *websocket.Conn
	
	// Send is a buffered channel for outbound messages.
	// Instead of writing directly to the websocket from multiple places (which crashes Go),
	// we push messages into this channel, and the WritePump handles the actual transmission.
	Send chan []byte
}

// ReadPump pumps messages from the websocket connection to the Room.
func (c *Client) ReadPump() {
	// Defer guarantees we clean up if the loop breaks or the client disconnects
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}
		
		// For now, we just pass the raw JSON byte array to the room's broadcast channel.
		// Later, we will unmarshal this JSON to check if it's "create_room", "input", etc.
		c.Room.Broadcast <- message
	}
}

// WritePump pumps messages from the Room back to the websocket connection.
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		// This blocks until a message is pushed into the c.Send channel
		message, ok := <-c.Send
		if !ok {
			// The room closed the channel.
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
	}
}