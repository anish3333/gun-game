package server

import (
	"io/fs"
	"net/http"
	"net/http/pprof"
)

func (s *Server) Run() error {

	mux := http.NewServeMux()

	// websocket
	mux.HandleFunc("/ws", s.HandleConnections)

	// api
	mux.HandleFunc("/api/init-guest", s.InitGuestHandler)

	// admin
	mux.HandleFunc("/admin/metrics", s.tracker.HandleAdminWS)
	mux.HandleFunc("/admin/strategy", s.AdminSwapStrategyHandler)

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// frontend
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return err
	}

	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	return http.ListenAndServe(":"+s.cfg.Port, mux)
}