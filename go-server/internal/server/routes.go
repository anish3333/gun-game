package server

import (
	"io/fs"
	"net/http"
	"net/http/pprof"
	"regexp"
	"strings"
)

var playRoomCodePattern = regexp.MustCompile(`^[A-Z0-9]{6}$`)

func (s *Server) Run() error {

	mux := http.NewServeMux()

	// websocket
	mux.HandleFunc("/ws", s.HandleConnections)

	// api
	mux.HandleFunc("/api/init-guest", s.InitGuestHandler)
	mux.HandleFunc("/api/room/", s.RoomInfoHandler)

	// admin
	mux.HandleFunc("/admin/metrics", s.tracker.HandleAdminWS)
	mux.HandleFunc("/admin/strategy", s.AdminSwapStrategyHandler)
	mux.HandleFunc("/admin/encoding", s.AdminSwapEncodingHandler)

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// frontend — SPA routes
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return err
	}

	mux.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		code := strings.TrimPrefix(r.URL.Path, "/play/")
		code = strings.Trim(code, "/")
		// Only serve the SPA shell for /play/{code} — not /play/js/... etc.
		if code == "" || strings.Contains(code, "/") || !playRoomCodePattern.MatchString(strings.ToUpper(code)) {
			http.NotFound(w, r)
			return
		}
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	return http.ListenAndServe(":"+s.cfg.Port, mux)
}