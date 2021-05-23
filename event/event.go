package event

import "encoding/json"

type Event struct {
	Type    string
	From    int
	Payload json.RawMessage
}

func ParseEvent(raw []byte) (e Event, err error) {
	err = json.Unmarshal(raw, &e)
	return
}
