package network

import "github.com/anish3333/gun-game/go-server/internal/game"

type IncomingMessage struct {
	Type     string   `json:"type"`
	Code     string   `json:"code,omitempty"`
	Weapon   string   `json:"weapon,omitempty"`
	Angle    *float64 `json:"angle,omitempty"`
	Shoot    bool     `json:"shoot,omitempty"`
	PlayerID string   `json:"playerId,omitempty"`

	// create_room / update_room_config
	ScoreLimit *int   `json:"scoreLimit,omitempty"`
	TimeLimit  *int   `json:"timeLimit,omitempty"`
	WeaponMode string `json:"weaponMode,omitempty"`
	Map        string `json:"map,omitempty"`
}

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
	InviteURL          string        `json:"inviteUrl,omitempty"`
	Reason             string        `json:"reason,omitempty"`
}

type SnapshotMessage struct {
	Type    string        `json:"type"`
	Players []game.Player `json:"players"`
	Bullets []game.Bullet `json:"bullets"`
}

type RoomStateMessage struct {
	Type           string        `json:"type"`
	Code           string        `json:"code"`
	Phase          string        `json:"phase"`
	HostID         string        `json:"hostId"`
	Players        []LobbyPlayer `json:"players"`
	Config         RoomConfig    `json:"config"`
	InviteURL      string        `json:"inviteUrl,omitempty"`
	SpectatorCount int           `json:"spectatorCount,omitempty"`
}

type RoomConfigUpdateMessage struct {
	Type   string     `json:"type"`
	Config RoomConfig `json:"config"`
}

type MatchResultsMessage struct {
	Type     string                      `json:"type"`
	WinnerID string                      `json:"winnerId"`
	Stats    map[string]PlayerMatchStats `json:"stats"`
}

type PlayerMatchStats struct {
	Kills  int `json:"kills"`
	Deaths int `json:"deaths"`
}

type HelloMessage struct {
	Type     string                    `json:"type"`
	Weapons  map[string]game.WeaponDef `json:"weapons"`
	Encoding string                    `json:"encoding"`
}

type EncodingChangedMessage struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
}

type RoomInfoResponse struct {
	Exists      bool   `json:"exists"`
	Code        string `json:"code,omitempty"`
	Phase       string `json:"phase,omitempty"`
	PlayerCount int    `json:"playerCount,omitempty"`
	CanJoin     bool   `json:"canJoin,omitempty"`
	CanSpectate bool   `json:"canSpectate,omitempty"`
}
