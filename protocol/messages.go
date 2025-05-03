package protocol

import (
	"errors"
	"net"
)

func init() {
	registerMessages([]Message{
		&BroadcastMessage{},
		&HostHeartbeatMessage{},
		&StateMessage{},
	})
}

type BroadcastMessage struct {
	Model    string
	Version  string
	IP       net.IP
	Mac      net.HardwareAddr
	Unknown2 []byte
}

func (bc *BroadcastMessage) ID() byte {
	return 0x01
}

func (bc *BroadcastMessage) EncodePayload() ([]byte, error) {
	var parts [][]byte
	parts = append(parts, []byte(bc.Version))
	parts = append(parts, []byte(bc.Model))
	parts = append(parts, []byte(bc.IP.String()))
	parts = append(parts, []byte(bc.Mac))
	parts = append(parts, bc.Unknown2)
	return encodeLV(parts)
}

func (bc *BroadcastMessage) DecodePayload(data []byte) error {
	parts, err := decodeLV(data)
	if err != nil {
		return err
	}
	if len(parts) != 5 {
		return errors.New("invalid number of parts")
	}
	bc.Version = string(parts[0])
	bc.Model = string(parts[1])
	bc.IP = net.ParseIP(string(parts[2]))
	if bc.IP == nil {
		return errors.New("invalid IP address")
	}
	bc.Mac = net.HardwareAddr(parts[3])
	bc.Unknown2 = parts[4]
	return nil
}

type ConnectRequestMessage struct {
	ClientIP net.IP
}

func (cr *ConnectRequestMessage) ID() byte {
	return 0x02
}

func (cr *ConnectRequestMessage) EncodePayload() ([]byte, error) {
	parts := [][]byte{[]byte(cr.ClientIP.String())}
	lv, err := encodeLV(parts)
	if err != nil {
		return nil, err
	}
	return append([]byte{0x00}, lv...), nil
}

func (cr *ConnectRequestMessage) DecodePayload(data []byte) error {
	return nil
}

type HostHeartbeatMessage struct {
}

func (hb *HostHeartbeatMessage) ID() byte {
	return 0x03
}
func (hb *HostHeartbeatMessage) EncodePayload() ([]byte, error) {
	return nil, nil
}
func (hb *HostHeartbeatMessage) DecodePayload(data []byte) error {
	return nil
}

type ClientHeartbeatMessage struct {
}

func (hb *ClientHeartbeatMessage) ID() byte {
	return 0x04
}
func (hb *ClientHeartbeatMessage) EncodePayload() ([]byte, error) {
	return nil, nil
}
func (hb *ClientHeartbeatMessage) DecodePayload(data []byte) error {
	return nil
}

type StateType byte

const (
	PowerState StateType = 0x01
	LightState StateType = 0x02
)

type SetStateMessage struct {
	Power *Power
	Light *Light
}

func (sm *SetStateMessage) ID() byte {
	return 0x05
}

func (sm *SetStateMessage) EncodePayload() ([]byte, error) {
	return encodeState(sm.Power, sm.Light)
}

func (sm *SetStateMessage) DecodePayload(data []byte) (err error) {
	sm.Power, sm.Light, err = decodeState(data)
	return
}

type QueryStateMessage struct {
	Type StateType
}

func (qs *QueryStateMessage) ID() byte {
	return 0x06
}

func (qs *QueryStateMessage) EncodePayload() ([]byte, error) {
	return []byte{byte(qs.Type)}, nil
}

func (qs *QueryStateMessage) DecodePayload(data []byte) error {
	if len(data) != 1 {
		return errors.New("invalid length")
	}
	qs.Type = StateType(data[0])
	return nil
}

type StateMessage struct {
	Power *Power
	Light *Light
}

func (sm *StateMessage) ID() byte {
	return 0x07
}

func (sm *StateMessage) EncodePayload() ([]byte, error) {
	return encodeState(sm.Power, sm.Light)
}

func (sm *StateMessage) DecodePayload(data []byte) (err error) {
	sm.Power, sm.Light, err = decodeState(data)
	return
}

func encodeState(p *Power, l *Light) ([]byte, error) {
	if p == nil && l == nil {
		return nil, errors.New("no state to encode")
	}
	if p != nil && l != nil {
		return nil, errors.New("both power and light states are set")
	}
	if p != nil {
		data, err := p.EncodePayload()
		if err != nil {
			return nil, err
		}
		return append([]byte{byte(PowerState)}, data...), nil
	}
	data, err := l.EncodePayload()
	if err != nil {
		return nil, err
	}
	return append([]byte{byte(LightState)}, data...), nil
}

func decodeState(data []byte) (*Power, *Light, error) {
	if len(data) < 2 {
		return nil, nil, errors.New("invalid length")
	}
	switch StateType(data[0]) {
	case PowerState:
		p := &Power{}
		err := p.DecodePayload(data[1:])
		if err != nil {
			return nil, nil, err
		}
		return p, nil, nil
	case LightState:
		l := &Light{}
		err := l.DecodePayload(data[1:])
		if err != nil {
			return nil, nil, err
		}
		return nil, l, nil
	default:
		return nil, nil, errors.New("unknown state type")
	}
}

type Power struct {
	On bool
}

func (p *Power) EncodePayload() ([]byte, error) {
	if p.On {
		return []byte{0x01}, nil
	}
	return []byte{0x00}, nil
}
func (p *Power) DecodePayload(data []byte) error {
	if len(data) != 1 {
		return errors.New("invalid length")
	}
	p.On = data[0] == 0x01
	return nil
}

type Light struct {
	Brightness  int
	Temperature int
}

func (l *Light) EncodePayload() ([]byte, error) {
	if l.Brightness < 0 || l.Brightness > 100 {
		return nil, errors.New("invalid brightness")
	}
	if l.Temperature < 29 || l.Temperature > 70 {
		return nil, errors.New("invalid temperature")
	}
	return []byte{byte(l.Brightness), byte(l.Temperature), 0x00}, nil
}

func (l *Light) DecodePayload(data []byte) error {
	if len(data) != 3 {
		return errors.New("invalid length")
	}
	l.Brightness = int(data[0])
	l.Temperature = int(data[1])
	return nil
}
