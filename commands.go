package neewer

import (
	"fmt"

	"github.com/anna-oake/go-neewer/protocol"
)

func (n *Neewer) SetPower(power bool) {
	n.txQueue <- &protocol.SetStateMessage{
		Power: &protocol.Power{
			On: power,
		},
	}
}

func (n *Neewer) SetLight(brightness, temperature int) error {
	if brightness < 0 || brightness > 100 {
		return fmt.Errorf("brightness must be between 0 and 100")
	}
	if temperature < 29 || temperature > 70 {
		return fmt.Errorf("temperature must be between 29 and 70")
	}
	n.txQueue <- &protocol.SetStateMessage{
		Light: &protocol.Light{
			Brightness:  brightness,
			Temperature: temperature,
		},
	}
	n.txQueue <- &protocol.QueryStateMessage{
		Type: protocol.LightState,
	}
	return nil
}
