package game

import (
	"bufio"
	"encoding/json"
	"io"
	"math"
	"math/rand"
)

type Controller interface {
	Tick(state *GameState, playerID string) InputState
}

type HumanController struct {
	Input InputState
}

func (h *HumanController) Tick(state *GameState, playerID string) InputState {
	return h.Input
}

type RandomController struct{}

func (r *RandomController) Tick(state *GameState, playerID string) InputState {
	angle := rand.Float64() * 2 * math.Pi
	return InputState{
		Angle: &angle,
		Shoot: rand.Float64() < 0.10,
	}
}

type PPOController struct {
	enc *json.Encoder
	dec *json.Decoder
}

func NewPPOController(stdin io.Writer, stdout io.Reader) *PPOController {
	return &PPOController{
		enc: json.NewEncoder(stdin),
		dec: json.NewDecoder(bufio.NewReader(stdout)),
	}
}

func (p *PPOController) Tick(state *GameState, playerID string) InputState {
	obs := BuildObservation(state, playerID)
	if err := p.enc.Encode(obs); err != nil {
		return InputState{}
	}

	var resp struct {
		Angle float64 `json:"angle"`
		Shoot bool    `json:"shoot"`
	}
	if err := p.dec.Decode(&resp); err != nil {
		return InputState{}
	}

	return InputState{
		Angle: &resp.Angle,
		Shoot: resp.Shoot,
	}
}
