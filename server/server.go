package server

import (
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

type Server struct {
	rooms map[string]*Room

	choose func(rooms map[string]*Room) (*Room, error)
}

func NewServer(choose func(rooms map[string]*Room) (*Room, error)) http.Handler {
	m := http.NewServeMux()
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(`./static`))))
	m.Handle("/socket", &Server{rooms: map[string]*Room{}, choose: choose})
	m.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "static/index.html")
	})
	return m
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("failed to create websocket connection: %v", err)
		return
	}
	defer func() {
		_ = c.Close(websocket.StatusInternalError, "something wrong happened")
	}()

	room, err := s.choose(s.rooms)
	if err != nil {
		log.Printf("failed to get room: %v", err)
		return
	}
	if err = room.Run(c); err != nil && websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		log.Printf("failed to run room: %v", err)
	}
}
