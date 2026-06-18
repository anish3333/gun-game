package network

import (
	"log"
	"strings"

	"github.com/anish3333/gun-game/go-server/internal/game"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID          string
	DisplayName string
	Weapon      string
	Manager     *Manager
	Room        *Room
	Conn        *websocket.Conn
	Send        chan Frame
}

func (c *Client) ReadPump() {
	defer func() {
		c.Manager.UnregisterClient(c)
		if c.Room != nil {
			c.Room.Unregister <- c
		}
		c.Conn.Close()
	}()

	c.Manager.RegisterClient(c)

	for {
		messageType, rawMessage, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		binary := messageType == websocket.BinaryMessage

		var msg IncomingMessage
		if err := decodeIncoming(binary, rawMessage, &msg); err != nil {
			continue
		}

		msg.PlayerID = c.ID

		switch msg.Type {
		case "ping":
			c.Send <- c.Manager.EncodeFrame(OutgoingMessage{Type: "pong"})

		case "create_room":
			if c.Room != nil {
				continue
			}
			c.Weapon = pickWeapon(msg.Weapon)
			config := parseRoomConfig(msg)

			room := c.Manager.CreateRoom(c.ID, config)
			room.Register <- c

			log.Printf("[%s] created by %s (%s)", room.ID, c.ID, c.Weapon)
			c.Send <- c.Manager.EncodeFrame(OutgoingMessage{
				Type:      "room_created",
				Code:      room.ID,
				PlayerID:  c.ID,
				Weapon:    c.Weapon,
				InviteURL: c.Manager.InviteURL(room.ID),
			})

		case "join_room":
			if c.Room != nil {
				continue
			}
			c.Weapon = pickWeapon(msg.Weapon)

			code := strings.ToUpper(strings.TrimSpace(msg.Code))
			room, exists := c.Manager.GetRoom(code)
			if !exists {
				c.sendError("Room not found.")
				continue
			}

			if exists, connected := room.PlayerStatus(c.ID); exists {
				if !connected {
					room.Register <- c
					c.Send <- c.Manager.EncodeFrame(OutgoingMessage{
						Type:     "room_joined",
						Code:     room.ID,
						PlayerID: c.ID,
						Weapon:   c.Weapon,
					})
					continue
				}
				c.sendError("You are already in this room.")
				continue
			}

			if !room.canJoin() {
				room.RegisterSpectator(c)
				continue
			}

			c.Room = room
			room.Register <- c

			log.Printf("[%s] joined by %s (%s)", room.ID, c.ID, c.Weapon)
			c.Send <- c.Manager.EncodeFrame(OutgoingMessage{
				Type:     "room_joined",
				Code:     room.ID,
				PlayerID: c.ID,
				Weapon:   c.Weapon,
			})

		case "spectate_room":
			if c.Room != nil {
				continue
			}

			code := strings.ToUpper(strings.TrimSpace(msg.Code))
			room, exists := c.Manager.GetRoom(code)
			if !exists {
				c.sendError("Room not found.")
				continue
			}

			if exists, connected := room.PlayerStatus(c.ID); exists && connected {
				c.sendError("You are already playing in this room.")
				continue
			}

			room.RegisterSpectator(c)

		case "leave_room":
			if c.Room == nil {
				continue
			}
			room := c.Room
			room.Leave <- c.ID

		case "start_match":
			if c.Room == nil {
				continue
			}
			c.Room.StartMatch <- c.ID

		case "update_room_config":
			if c.Room == nil || c.Room.HostID != c.ID {
				continue
			}
			config := c.Room.Config
			if msg.ScoreLimit != nil {
				config.ScoreLimit = clampInt(*msg.ScoreLimit, 1, 50)
			}
			if msg.TimeLimit != nil {
				config.TimeLimit = clampInt(*msg.TimeLimit, 1, 30)
			}
			if msg.WeaponMode != "" {
				config.WeaponMode = msg.WeaponMode
			}
			if msg.Map != "" {
				config.Map = msg.Map
			}
			c.Room.UpdateConfig <- config

		case "rematch":
			if c.Room == nil {
				continue
			}
			c.Room.Rematch <- c.ID

		case "input":
			if c.Room != nil {
				c.Room.Broadcast <- msg
			}
		}
	}
}

func (c *Client) sendError(message string) {
	c.Send <- c.Manager.EncodeFrame(OutgoingMessage{Type: "error", Message: message})
}

func parseRoomConfig(msg IncomingMessage) RoomConfig {
	config := DefaultRoomConfig()
	if msg.ScoreLimit != nil {
		config.ScoreLimit = clampInt(*msg.ScoreLimit, 1, 50)
	}
	if msg.TimeLimit != nil {
		config.TimeLimit = clampInt(*msg.TimeLimit, 1, 30)
	}
	if msg.WeaponMode != "" {
		config.WeaponMode = msg.WeaponMode
	}
	if msg.Map != "" {
		config.Map = msg.Map
	}
	return config
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func pickWeapon(weapon string) string {
	if weapon != "" {
		if _, ok := game.Weapons[weapon]; ok {
			return weapon
		}
	}
	return "pistol"
}

func BuildHello(m *Manager) Frame {
	return m.EncodeFrame(HelloMessage{
		Type:     "hello",
		Weapons:  game.Weapons,
		Encoding: m.GetCodecName(),
	})
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for {
		frame, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		msgType := websocket.TextMessage
		if frame.Binary {
			msgType = websocket.BinaryMessage
		}
		c.Conn.WriteMessage(msgType, frame.Payload)
	}
}
