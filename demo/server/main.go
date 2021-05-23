package main

import (
	"log"
	"net/http"

	"github.com/arovesto/gio/demo/entities"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/event"
	"github.com/arovesto/gio/math"
	"github.com/arovesto/gio/misc"
	"github.com/arovesto/gio/server"
)

var lobby = server.NewBasicRoom(0, "lobby", []elements.Element{
	&elements.StaticBackground{
		Where:     math.Box{Corner: math.Vector{X: 100, Y: 100}, Size: math.Vector{X: 1464, Y: 720}},
		TextureID: "lobby.png",
		ID:        0,
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
	server.CustomRoomAction["lobby"] = func(r *server.Room) {
		p := r.Players()
		if len(p) >= 2 {
			room := server.NewBasicRoom(misc.NewID(), "snake", []elements.Element{
				&entities.Snake{
					Orbs: []math.Sphere{{Center: math.Vector{X: 200, Y: 200}, R: 50}},
					ID:   10,
				},
				&entities.Snake{
					Orbs: []math.Sphere{{Center: math.Vector{X: 1200, Y: 700}, R: 50}},
					ID:   20,
				},
			})
			if err := r.Transfer(p[0], room); err != nil {
				log.Println("failed to transfer guy 1 to new room", err)
			}
			if err := r.Transfer(p[1], room); err != nil {
				log.Println("failed to transfer guy 1 to new room", err)
			}
		}
	}
	server.CustomRoomAction["snake"] = func(r *server.Room) {
		st, ok := r.CustomState.(bool)
		p := r.Players()
		if len(p) == 1 && ok && st {
			if err := r.Transfer(p[0], winLobby); err != nil {
				log.Println("failed to transfer won player to win room")
			}
		}
	}
	server.EventsProcessors["snake"] = map[string]func(e event.Event, r *server.Room) error{
		"lose": func(e event.Event, r *server.Room) error {
			r.CustomState = true
			return r.Transfer(e.From, loseLobby)
		},
	}
	server.PlayerChoiceFunctions["lobby"] = func(playable map[int]elements.Playable, assigned map[int]struct{}, r *server.Room) (int, error) {
		id := r.NewID()
		r.NewElement(&elements.NoOpPlayer{ID: id})
		return id, nil
	}
	server.PlayerChoiceFunctions["game-over-lobby"] = func(playable map[int]elements.Playable, assigned map[int]struct{}, r *server.Room) (int, error) {
		id := r.NewID()
		r.NewElement(&entities.GameOverPlayer{NoOpPlayer: elements.NoOpPlayer{ID: id}, Lobby: lobby})
		return id, nil
	}
	server.PlayerChoiceFunctions["snake"] = func(playable map[int]elements.Playable, assigned map[int]struct{}, r *server.Room) (int, error) {
		if _, ok := assigned[10]; !ok {
			return 10, nil
		}
		if _, ok := assigned[20]; !ok {
			return 20, nil
		}
		return 0, server.RoomFull
	}

	if err := http.ListenAndServe(`:8080`, server.NewServer(func(rooms map[string]*server.Room) (*server.Room, error) {
		return lobby, nil
	})); err != http.ErrServerClosed {
		panic(err)
	}
}
