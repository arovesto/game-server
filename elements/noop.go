package elements

import (
	"encoding/json"

	"github.com/arovesto/gio/canvas"
)

type NoOpPlayer struct {
	ID int
}

func (n *NoOpPlayer) Draw(c canvas.Canvas) {
	return
}

func (n *NoOpPlayer) GetID() int {
	return n.ID
}

func (n *NoOpPlayer) GetType() int {
	return NoOpPlayerType
}

func (n *NoOpPlayer) GetState() ([]byte, error) {
	return json.Marshal(n)
}

func (n *NoOpPlayer) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, n)
}

func (n *NoOpPlayer) Input() ([]byte, error) {
	return nil, nil
}

func (n *NoOpPlayer) SetInput(bytes []byte) error {
	return nil
}
