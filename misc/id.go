package misc

import (
	"sync"
)

var mtx sync.Mutex
var curID int

func NewID() int {
	mtx.Lock()
	defer mtx.Unlock()
	c := curID
	curID++
	return c
}
