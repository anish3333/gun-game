package network

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/anish3333/gun-game/go-server/internal/game"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID      string
	Weapon  string
	Manager *Manager
	Room    *Room
	Conn    *websocket.Conn
	Send    chan []byte
}

func (c *Client) ReadPump() {
	defer func() {
		if c.Room != nil { c.Room.Unregister <- c }
		c.Conn.Close()
	}()

	for {
		_, rawMessage, err := c.Conn.ReadMessage()
		if err != nil { break }

		var msg IncomingMessage
		if err := json.Unmarshal(rawMessage, &msg); err != nil { continue }
		
		// Inject the Client ID into the message so the Room knows who sent it
		msg.PlayerID = c.ID 

		switch msg.Type {
		case "ping":
			pong, _ := json.Marshal(OutgoingMessage{Type: "pong"})
			c.Send <- pong

		case "list_rooms":
			res, _ := json.Marshal(RoomListMessage{Type: "room_list", Rooms: c.Manager.RoomList()})
			c.Send <- res

		case "create_room":
			if c.Room != nil { continue }
			c.Weapon = pickWeapon(msg.Weapon)

			c.Room = c.Manager.CreateRoom()
			c.Room.Register <- c

			log.Printf("[%s] created by %s (%s)", c.Room.ID, c.ID, c.Weapon)
			res, _ := json.Marshal(OutgoingMessage{Type: "room_created", Code: c.Room.ID, PlayerID: c.ID, Weapon: c.Weapon, WaitingForOpponent: true})
			c.Send <- res

		case "join_room":
			if c.Room != nil { continue }
			c.Weapon = pickWeapon(msg.Weapon)

			code := strings.ToUpper(strings.TrimSpace(msg.Code))
			room, exists := c.Manager.GetRoom(code)
			if !exists {
				res, _ := json.Marshal(OutgoingMessage{Type: "error", Message: "Room not found."})
				c.Send <- res
				continue
			}
			if len(room.Clients) >= 2 {
				res, _ := json.Marshal(OutgoingMessage{Type: "error", Message: "Room is full."})
				c.Send <- res
				continue
			}
			if room.Phase != "waiting" {
				res, _ := json.Marshal(OutgoingMessage{Type: "error", Message: "Match already started."})
				c.Send <- res
				continue
			}

			c.Room = room
			c.Room.Register <- c

			log.Printf("[%s] joined by %s (%s)", room.ID, c.ID, c.Weapon)
			res, _ := json.Marshal(OutgoingMessage{Type: "room_joined", Code: room.ID, PlayerID: c.ID, Weapon: c.Weapon})
			c.Send <- res

		case "input":
			if c.Room != nil {
				b, _ := json.Marshal(IncomingMessage{
					Type:     "input",
					PlayerID: c.ID,
					Angle:    msg.Angle,
					Shoot:    msg.Shoot,
				})
				c.Room.Broadcast <- b
			}
		}
	}
}

func pickWeapon(weapon string) string {
	if weapon != "" {
		if _, ok := game.Weapons[weapon]; ok {
			return weapon
		}
	}
	return "pistol"
}

func BuildHello(m *Manager) []byte {
	b, _ := json.Marshal(HelloMessage{
		Type:    "hello",
		Weapons: game.Weapons,
		Rooms:   m.RoomList(),
	})
	return b
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for {
		message, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		c.Conn.WriteMessage(websocket.TextMessage, message)
	}
}