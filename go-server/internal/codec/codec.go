package codec

import (
	"bytes"
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	JSON    = "json"
	MsgPack = "msgpack"
)

type Codec interface {
	Name() string
	IsBinary() bool
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

func New(name string) Codec {
	switch name {
	case MsgPack:
		return &MsgPackCodec{}
	default:
		return &JSONCodec{}
	}
}



type JSONCodec struct{}

func (JSONCodec) Name() string    { return JSON }
func (JSONCodec) IsBinary() bool  { return false }

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}



type MsgPackCodec struct{}

func (MsgPackCodec) Name() string   { return MsgPack }
func (MsgPackCodec) IsBinary() bool { return true }

func (MsgPackCodec) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (MsgPackCodec) Unmarshal(data []byte, v interface{}) error {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.SetCustomStructTag("json")
	return dec.Decode(v)
}

// DecodeFrame unmarshals based on WebSocket frame type (text=json, binary=msgpack).
func DecodeFrame(binary bool, data []byte, v interface{}) error {
	if binary {
		return MsgPackCodec{}.Unmarshal(data, v)
	}
	return JSONCodec{}.Unmarshal(data, v)
}
