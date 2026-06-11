package server

import (
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/anish3333/gun-game/go-server/internal/config"
	"github.com/anish3333/gun-game/go-server/internal/db"
	"github.com/anish3333/gun-game/go-server/internal/network"
	"github.com/anish3333/gun-game/go-server/internal/telemetry"
	
)

type Server struct {
	cfg     *config.Config
	db      *db.Database
	manager *network.Manager
	tracker *telemetry.Tracker
	upgrader websocket.Upgrader
}

func New(
	cfg *config.Config,
	db *db.Database,
	manager *network.Manager,
	tracker *telemetry.Tracker,
) *Server {

	return &Server{
		cfg: cfg,
		db: db,
		manager: manager,
		tracker: tracker,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}