package neewer

import (
	"log"
	"net"
	"time"

	"github.com/anna-oake/go-neewer/protocol"
)

func (n *Neewer) connect() error {
	if n.tx != nil {
		n.tx.Close()
		n.tx = nil
	}
	if n.rx != nil {
		n.rx.Close()
		n.rx = nil
	}
	n.txQueue = make(chan protocol.Message, 10)
	n.rxQueue = make(chan protocol.Message, 10)

	var err error

	n.tx, err = net.DialUDP("udp", nil, n.hostAddr)
	if err != nil {
		return err
	}

	n.rx, err = net.ListenUDP("udp4", n.clientAddr)
	if err != nil {
		return err
	}

	go n.rxLoop()
	go n.txLoop()
	go n.heartbeatLoop()
	go n.receive()
	n.sendHeartbeat()

	return nil
}

func (n *Neewer) rxLoop() {
	var buffer [512]byte
	for {
		num, src, err := n.rx.ReadFromUDP(buffer[:])
		if err != nil {
			log.Println("Error reading from UDP:", err)
			continue
		}

		if !src.IP.Equal(n.hostAddr.IP) {
			log.Println("Received message from unexpected source:", src)
			continue
		}

		data := buffer[:num]

		for len(data) > 0 {
			msg, rest, err := protocol.Parse(data)
			if err != nil {
				log.Println("Error parsing message:", err)
			} else {
				n.rxQueue <- msg
				data = rest
			}
		}
	}
}

func (n *Neewer) txLoop() {
	for msg := range n.txQueue {
		b, err := protocol.Serialize(msg)
		if err != nil {
			log.Println("Error serializing message:", err)
			continue
		}

		_, err = n.tx.Write(b)
		if err != nil {
			log.Println("Error writing to UDP:", err)
			n.txQueue <- msg
			continue
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (n *Neewer) heartbeatLoop() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for range t.C {
		n.sendHeartbeat()
	}
}

func (n *Neewer) sendHeartbeat() {
	if n.State.Alive && time.Since(n.lastHeartbeat) > 5*time.Second {
		n.State.Alive = false
		if n.OnAliveChange != nil {
			n.OnAliveChange(false)
		}
	}
	if !n.State.Alive {
		n.txQueue <- &protocol.ConnectRequestMessage{
			ClientIP: n.clientAddr.IP,
		}
	} else {
		n.txQueue <- &protocol.ClientHeartbeatMessage{}
		n.txQueue <- &protocol.QueryStateMessage{
			Type: protocol.LightState,
		}
	}
}
