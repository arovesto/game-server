package gio

import "github.com/arovesto/gio/math"

// add to this map your element types
var GenElements = map[string]func()Element{}

// should be json + yaml marshalled to allow sending to web + storing in config files
type Element interface {
	Draw(c *Canvas)         // all actions to draw the object
	Update() error               // all changes in object
	Collide(other Element) error // collision should be checked inside
	Collider() math.Shape        // useful in Collide
}

