package server

import (
	"log"
	"net/http"

	"github.com/anish3333/gun-game/go-server/internal/codec"
	"github.com/anish3333/gun-game/go-server/internal/engine"
)

func (s *Server) AdminSwapStrategyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	strategy := r.URL.Query().Get("type")

	if strategy == "spatial" {
		s.manager.SetCollisionEngine(engine.NewSpatialHashEngine(), "SpatialHash O(N)")
		log.Println("✦ Admin Hot-Swapped Engine -> SpatialHash")
	} else {
		s.manager.SetCollisionEngine(engine.NewBruteForceEngine(), "BruteForce O(N^2)")
		log.Println("✦ Admin Hot-Swapped Engine -> BruteForce")
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) AdminSwapEncodingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	encoding := r.URL.Query().Get("type")
	switch encoding {
	case codec.MsgPack:
		s.manager.SetCodec(codec.MsgPack)
		log.Println("✦ Admin Hot-Swapped Encoding -> msgpack")
	case codec.JSON:
		s.manager.SetCodec(codec.JSON)
		log.Println("✦ Admin Hot-Swapped Encoding -> json")
	default:
		http.Error(w, "type must be json or msgpack", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}