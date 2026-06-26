package network

import "time"

type RoomPhase int

const (
	PhaseLobby RoomPhase = iota
	PhaseCountdown
	PhasePlaying
	PhaseFinished
)

func (p RoomPhase) String() string {
	switch p {
	case PhaseLobby:
		return "lobby"
	case PhaseCountdown:
		return "countdown"
	case PhasePlaying:
		return "playing"
	case PhaseFinished:
		return "finished"
	default:
		return "unknown"
	}
}

type RoomConfig struct {
	ScoreLimit int    `json:"scoreLimit"`
	TimeLimit  int    `json:"timeLimit"`
	WeaponMode string `json:"weaponMode"`
	Map        string `json:"map"`
	BotEnabled bool   `json:"botEnabled"`
}

func DefaultRoomConfig() RoomConfig {
	return RoomConfig{
		ScoreLimit: 10,
		TimeLimit:  5,
		WeaponMode: "any",
		Map:        "arena",
	}
}

type LobbyPlayer struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Weapon            string    `json:"weapon"`
	IsHost            bool      `json:"isHost"`
	Connected         bool      `json:"connected"`
	RematchVote       bool      `json:"rematchVote,omitempty"`
	ReconnectSeconds  int       `json:"reconnectSeconds,omitempty"`
	ReconnectDeadline time.Time `json:"-"`
}
