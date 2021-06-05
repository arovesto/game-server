package main

import (
	"net/http"

	"github.com/arovesto/gio/demo/entities"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/event"
	"github.com/arovesto/gio/math"
	"github.com/arovesto/gio/server"
)

// TODO add walls
var lobby = server.NewBasicRoom(0, "lobby", []elements.Element{
	&elements.StaticBackground{
		Where:     math.Box{Corner: math.Vector{X: 100, Y: 100}, Size: math.Vector{X: 1464, Y: 720}},
		TextureID: "lobby.png",
		ID:        0,
	},
	&entities.Trigger{
		ID:     1,
		Gather: math.Box{Corner: math.Vector{X: 1000, Y: 1000}, Size: math.Vector{X: 500, Y: 500}},
		Start:  math.Box{Corner: math.Vector{X: 1000, Y: 1000}, Size: math.Vector{X: 200, Y: 200}},
		Ready:  map[int]struct{}{},
	},
})

var loseLobby = server.NewBasicRoom(0, "game-over-lobby", []elements.Element{
	&elements.StaticBackground{
		Where:     math.Box{Corner: math.Vector{X: 100, Y: 100}, Size: math.Vector{X: 1464, Y: 720}},
		TextureID: "lose.png",
		ID:        0,
	},
})

var winLobby = server.NewBasicRoom(0, "game-over-lobby", []elements.Element{
	&elements.StaticBackground{
		Where:     math.Box{Corner: math.Vector{X: 100, Y: 100}, Size: math.Vector{X: 1464, Y: 720}},
		TextureID: "win.png",
		ID:        0,
	},
})

func main() {
	// TODO EventProcessor should be a Element interface with Subscribtions() map[string]struct{} and Process(e Event) error or smth
	// TODO after creating a "Room storage" transfer will use well-known id, so no need in "lose"
	// TODO DELETE ROOM WHEN NO PLAYERS PRESENT
	server.EventsProcessors["snake"] = map[string]func(e event.Event, r *server.Room) error{
		"lose": func(e event.Event, r *server.Room) error {
			return r.Transfer(e.From, loseLobby)
		},
		"win": func(e event.Event, r *server.Room) error {
			for _, p := range r.Players() {
				if err := r.Transfer(p, winLobby); err != nil {
					return err
				}
			}
			return nil
		},
	}
	server.PlayerChoiceFunctions["lobby"] = func(playable map[int]elements.Playable, assigned map[int]struct{}, r *server.Room) (int, error) {
		id := r.NewID()
		r.NewElement(entities.NewGuy(id))
		return id, nil
	}
	server.PlayerChoiceFunctions["game-over-lobby"] = func(playable map[int]elements.Playable, assigned map[int]struct{}, r *server.Room) (int, error) {
		id := r.NewID()
		r.NewElement(&entities.GameOverPlayer{NoOpPlayer: elements.NoOpPlayer{ID: id}, Lobby: lobby})
		return id, nil
	}
	server.PlayerChoiceFunctions["snake"] = func(playable map[int]elements.Playable, assigned map[int]struct{}, r *server.Room) (int, error) {
		for id := range playable {
			if _, ok := assigned[id]; !ok {
				return id, nil
			}
		}
		return 0, server.RoomFull
	}

	if err := http.ListenAndServe(`:8080`, server.NewServer(func(rooms map[string]*server.Room) (*server.Room, error) {
		return lobby, nil
	})); err != http.ErrServerClosed {
		panic(err)
	}
}
