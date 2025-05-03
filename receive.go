package neewer

import (
	"log"
	"time"

	"github.com/anna-oake/go-neewer/protocol"
)

func (n *Neewer) receive() {
	for msg := range n.rxQueue {
		if _, ok := msg.(*protocol.HostHeartbeatMessage); ok {
			n.lastHeartbeat = time.Now()
			if !n.State.Alive {
				n.State.Alive = true
				n.txQueue <- &protocol.QueryStateMessage{
					Type: protocol.PowerState,
				}
				if n.OnAliveChange != nil {
					n.OnAliveChange(true)
				}
				if n.OnStateChange != nil {
					n.OnStateChange(n.State)
				}
			}
			continue
		}

		switch msg := msg.(type) {
		case *protocol.GenericMessage:
			log.Printf("Received unknown message: cmd=%d, data=%x", msg.CmdID, msg.Raw)
		case *protocol.StateMessage:
			var changed bool
			if msg.Power != nil && n.State.Power != msg.Power.On {
				n.State.Power = msg.Power.On
				changed = true
				if n.OnPowerChange != nil {
					n.OnPowerChange(msg.Power.On)
				}
			}
			if msg.Light != nil {
				if n.State.Brightness != msg.Light.Brightness {
					n.State.Brightness = msg.Light.Brightness
					changed = true
					if n.OnBrightnessChange != nil {
						n.OnBrightnessChange(msg.Light.Brightness)
					}
				}
				if n.State.Temperature != msg.Light.Temperature {
					n.State.Temperature = msg.Light.Temperature
					changed = true
					if n.OnTemperatureChange != nil {
						n.OnTemperatureChange(msg.Light.Temperature)
					}
				}
			}
			if changed && n.OnStateChange != nil {
				n.OnStateChange(n.State)
			}
		}
	}
}
