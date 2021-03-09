package gio

type Runner struct {
	// all unexported, our staff (websocket etc)
}

type RunFunc func(r Room) error

type RoomInitFunc func() (Room, error)

func NewRunner(conf Config, f RunFunc) *Runner {
	return &Runner{}
}

// runs server, returns when stopped by signal or from console
func (r *Runner) Start() error {
	return nil
}
