package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) RoomInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	code := strings.ToUpper(strings.TrimPrefix(r.URL.Path, "/api/room/"))
	code = strings.Trim(code, "/")

	info := s.manager.RoomInfo(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
