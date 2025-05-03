package neewer

import (
	"net"
	"time"

	"github.com/anna-oake/go-neewer/protocol"
)

type Neewer struct {
	hostAddr   *net.UDPAddr
	clientAddr *net.UDPAddr

	txQueue       chan protocol.Message
	rxQueue       chan protocol.Message
	tx            *net.UDPConn
	rx            *net.UDPConn
	lastHeartbeat time.Time

	OnAliveChange       func(alive bool)
	OnPowerChange       func(power bool)
	OnBrightnessChange  func(brightness int)
	OnTemperatureChange func(temperature int)
	OnStateChange       func(state State)
	State               State
}

type State struct {
	Alive       bool
	Power       bool
	Brightness  int
	Temperature int
}

func NewNeewer(address string, clientIP string) (*Neewer, error) {
	hostAddr, err := net.ResolveUDPAddr("udp4", address+":5052")
	if err != nil {
		return nil, err
	}
	clientAddr, err := net.ResolveUDPAddr("udp4", clientIP+":5052")
	if err != nil {
		return nil, err
	}

	n := &Neewer{
		hostAddr:   hostAddr,
		clientAddr: clientAddr,
	}
	if err := n.connect(); err != nil {
		return nil, err
	}
	return n, nil
}
