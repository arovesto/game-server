package server

var id int

func newID() int {
	id++
	return id
}
