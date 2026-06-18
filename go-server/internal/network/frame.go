package network

import (
	"encoding/json"

	"github.com/anish3333/gun-game/go-server/internal/codec"
)

// Frame is a wire-ready WebSocket payload plus its frame type.
type Frame struct {
	Payload []byte
	Binary  bool // false=text/JSON, true=binary/msgpack
}

func (m *Manager) EncodeFrame(v interface{}) Frame {
	m.mu.RLock()
	c := m.Codec
	m.mu.RUnlock()

	payload, err := c.Marshal(v)
	if err != nil {
		return Frame{}
	}
	return Frame{Payload: payload, Binary: c.IsBinary()}
}

// EncodeControlJSON always emits a text/JSON frame (used for encoding_changed).
func (m *Manager) EncodeControlJSON(v interface{}) Frame {
	payload, _ := json.Marshal(v)
	return Frame{Payload: payload, Binary: false}
}

func decodeIncoming(binary bool, data []byte, v interface{}) error {
	return codec.DecodeFrame(binary, data, v)
}
