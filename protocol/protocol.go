package protocol

import (
	"errors"
	"reflect"
)

var factoryMap = make(map[byte]func() Message)

const frameHeader byte = 0x80

type frame struct {
	ID      byte
	Payload []byte
}

func parseFrame(data []byte) (*frame, []byte, error) {
	if len(data) < 4 {
		return nil, nil, errors.New("packet too short")
	}
	if data[0] != frameHeader {
		return nil, nil, errors.New("invalid header byte")
	}
	cmd := data[1]
	length := int(data[2])
	if 3+length+1 > len(data) {
		return nil, nil, errors.New("packet too short")
	}
	// checksum is last byte
	if computeChecksum(data[:3+length]) != data[3+length] {
		return nil, nil, errors.New("checksum mismatch")
	}
	// slice out payload
	payload := make([]byte, length)
	copy(payload, data[3:3+length])
	return &frame{ID: cmd, Payload: payload}, data[3+length+1:], nil
}

func (rp *frame) serialize() ([]byte, error) {
	if len(rp.Payload) > 255 {
		return nil, errors.New("payload too large")
	}
	buf := []byte{frameHeader, rp.ID, byte(len(rp.Payload))}
	buf = append(buf, rp.Payload...)
	buf = append(buf, computeChecksum(buf))
	return buf, nil
}

type Message interface {
	ID() byte
	EncodePayload() ([]byte, error)
	DecodePayload([]byte) error
}

func Serialize(p Message) ([]byte, error) {
	pl, err := p.EncodePayload()
	if err != nil {
		return nil, err
	}
	rp := &frame{ID: p.ID(), Payload: pl}
	return rp.serialize()
}

// Parse takes raw bytes, unwraps framing, and returns a TypedPacket
func Parse(data []byte) (Message, []byte, error) {
	rp, rest, err := parseFrame(data)
	if err != nil {
		return nil, rest, err
	}

	var cmd Message

	factory, ok := factoryMap[rp.ID]
	if ok {
		cmd = factory()
	} else {
		cmd = &GenericMessage{CmdID: rp.ID}
	}

	if err := cmd.DecodePayload(rp.Payload); err != nil {
		return nil, rest, err
	}

	return cmd, rest, nil
}

type GenericMessage struct {
	CmdID byte
	Raw   []byte
}

func (g *GenericMessage) ID() byte { return g.CmdID }

func (g *GenericMessage) EncodePayload() ([]byte, error) { return g.Raw, nil }

func (g *GenericMessage) DecodePayload(data []byte) error { g.Raw = data; return nil }

func registerMessage(c Message) {
	t := reflect.TypeOf(c).Elem()

	factoryMap[c.ID()] = func() Message {
		return reflect.New(t).Interface().(Message)
	}
}

func registerMessages(cmds []Message) {
	for _, cmd := range cmds {
		registerMessage(cmd)
	}
}
