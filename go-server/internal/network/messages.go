package network

import "github.com/anish3333/gun-game/go-server/internal/game"

type IncomingMessage struct {
	Type     string   `json:"type"`
	Code     string   `json:"code,omitempty"`
	Weapon   string   `json:"weapon,omitempty"`
	Angle    *float64 `json:"angle,omitempty"`
	Shoot    bool     `json:"shoot,omitempty"`
	
	PlayerID string   `json:"playerId,omitempty"`
}

// OutgoingMessage covers all the random JSON shapes your server sends
type OutgoingMessage struct {
	Type               string        `json:"type"`
	Message            string        `json:"message,omitempty"`
	Code               string        `json:"code,omitempty"`
	PlayerID           string        `json:"playerId,omitempty"`
	KillerID           string        `json:"killerId,omitempty"`
	WinnerID           string        `json:"winnerId,omitempty"` 
	Weapon             string        `json:"weapon,omitempty"`
	WaitingForOpponent bool          `json:"waitingForOpponent,omitempty"`
	Damage             int           `json:"damage,omitempty"`
	HP                 int           `json:"hp,omitempty"`
	Players            []game.Player `json:"players,omitempty"`
}

// SnapshotMessage perfectly matches your JS `buildSnapshot` output
type SnapshotMessage struct {
	Type    string        `json:"type"`
	Players []game.Player `json:"players"`
	Bullets []game.Bullet `json:"bullets"`
}

type RoomSummary struct {
	Code    string `json:"code"`
	Players int    `json:"players"`
}

type HelloMessage struct {
	Type    string                    `json:"type"`
	Weapons map[string]game.WeaponDef `json:"weapons"`
	Rooms   []RoomSummary             `json:"rooms"`
}

type RoomListMessage struct {
	Type  string        `json:"type"`
	Rooms []RoomSummary `json:"rooms"`
}