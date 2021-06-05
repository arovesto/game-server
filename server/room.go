package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"nhooyr.io/websocket"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/event"
)

var (
	EntityNotFound = errors.New("entity not found")
	RoomFull       = errors.New("room full")
)

const (
	layers  = 10
	Stopped = iota
	Running
)

var readAtMostEvents = 100

// TODO replace this with function field of a room, it is seams better
var PlayerChoiceFunctions = map[string]func(playable map[int]elements.Playable, assigned map[int]struct{}, r *Room) (int, error){} // what element of room should be used on new connection

// TODO replace this with local map in Room thing
var EventsProcessors = map[string]map[string]func(e event.Event, r *Room) error{} // use this map as event blablabla

type player struct {
	transfer chan *Room
	c        *websocket.Conn
}

type RawElement struct {
	Type int             `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Room struct {
	clients map[int]player
	events  chan event.Event
	done    chan struct{}

	State       int          `json:"state"`
	ID          int          `json:"id"`
	Type        string       `json:"type"`
	RawElements []RawElement `json:"elements"`
	CustomState interface{}  `json:"custom_state"`

	// TODO introduce da LAYER, so no z-fighting would occure
	elements    map[int]elements.Element
	players     map[int]elements.Playable
	movable     map[int]elements.Movable
	collidable  map[int]elements.Collidable
	drawOrder   []map[int]elements.Drawable
	toDelete    map[int]struct{}
	oneTickDiff map[int][]byte

	currentID int
}

func NewBasicRoom(id int, tp string, elms []elements.Element) *Room {
	r := &Room{}
	r.init(id, tp, elms)
	return r
}

func (s *Room) init(id int, tp string, elms []elements.Element) {
	s.elements = map[int]elements.Element{}
	s.movable = map[int]elements.Movable{}
	s.players = map[int]elements.Playable{}
	s.oneTickDiff = map[int][]byte{}
	s.collidable = map[int]elements.Collidable{}
	s.toDelete = map[int]struct{}{}
	s.drawOrder = make([]map[int]elements.Drawable, layers)
	s.ID = id
	s.Type = tp
	s.done = make(chan struct{})
	s.events = make(chan event.Event, readAtMostEvents)
	s.clients = map[int]player{}

	for _, el := range elms {
		s.NewElement(el)
	}
}

func (s *Room) GetElement(id int) elements.Element {
	return s.elements[id]
}

func (s *Room) GetElements() map[int]elements.Element {
	return s.elements
}

func (s *Room) Players() (r []int) {
	for id := range s.clients {
		r = append(r, id)
	}
	return
}

func (s *Room) Start() {
	go func() {
		now := time.Now()
		lastUpdate := now
		for {
			// TODO separate processEvents from Update, so all delta became consistent
			n := time.Now()
			if n.Sub(lastUpdate) > 16*time.Millisecond {
				st := time.Now()
				s.Update(n.Sub(lastUpdate))
				s.processEvents()
				lastUpdate = n
				if time.Since(st) > 16*time.Millisecond {
					log.Println("overload! main cycle can't keep up", time.Since(st))
				}
			}
			time.Sleep(time.Millisecond)
			now = n
		}
	}()

	s.State = Running
}

func (s *Room) Update(delta time.Duration) {
	s.toDelete = map[int]struct{}{}
	s.oneTickDiff = map[int][]byte{}

	for _, e := range s.movable {
		st, err := e.GetState()
		if err != nil {
			log.Println("error on get state", e.GetID(), e.GetType(), err)
			continue
		}
		s.oneTickDiff[e.GetID()] = st
		if err := e.Move(delta, s); err != nil {
			log.Println("error on move entity", e.GetID(), e.GetType(), err)
		}
	}

	for i1, e1 := range s.collidable {
		for i2, e2 := range s.collidable {
			if i1 != i2 {
				if err := e1.Collide(e2); err != nil {
					log.Println("collide error", err)
				}
			}
		}
	}
}

func (s *Room) DeleteElement(id int) {
	e := s.GetElement(id)

	delete(s.movable, id)
	delete(s.collidable, id)
	delete(s.elements, id)
	delete(s.players, id)
	delete(s.drawOrder[getElementLayer(e)], id)
}

func (s *Room) processEvents() {
	for i := 0; i < readAtMostEvents; i++ {
		select {
		case e := <-s.events:
			if err := s.ProcessEvent(e); err != nil {
				log.Println("failed to process event", err)
			}
		default:
			break
		}
	}
	for _, e := range s.movable {
		if _, ok := s.toDelete[e.GetID()]; ok {
			s.DeleteElement(e.GetID())
			s.BroadcastEvent(event.Event{Type: "deleted", From: e.GetID()})
			if p, ok := s.clients[e.GetID()]; ok {
				data, err := json.Marshal(event.Event{Type: "game-over", From: e.GetID()})
				if err != nil {
					log.Println("failed to mashal game over event", err)
				}
				if err := p.c.Write(context.TODO(), websocket.MessageText, data); err != nil {
					log.Println("failed to send game over event", err)
				}
			}
		}
		state, err := e.GetState()
		if err != nil {
			log.Println("error getting state", e.GetID(), e.GetType(), err, e)
		}
		if old, ok := s.oneTickDiff[e.GetID()]; ok && !bytes.Equal(old, state) {
			s.BroadcastEvent(event.Event{Type: "update", From: e.GetID(), Payload: state})
		}
	}
}

func (s *Room) Stop() {
	panic("implement me")
}

func (s *Room) GetState() ([]byte, error) {
	s.RawElements = s.RawElements[:0]

	for _, e := range s.elements {
		st, err := e.GetState()
		if err != nil {
			return nil, err
		}
		s.RawElements = append(s.RawElements, RawElement{
			Type: e.GetType(),
			Data: st,
		})
	}
	return json.Marshal(s)
}

func (s *Room) SetState(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	var elems = make([]elements.Element, 0, len(s.RawElements))
	for _, r := range s.RawElements {
		if f, ok := elements.GenElements[r.Type]; ok {
			e := f()
			if err := e.SetState(r.Data); err != nil {
				return err
			}
			elems = append(elems, e)
		} else {
			return fmt.Errorf("failed to locate entity type %d", r.Type)
		}
	}
	s.init(s.ID, s.Type, elems)
	return nil
}

func (s *Room) Run(c *websocket.Conn) error {
	if s.State != Running {
		s.Start()
	}
	ctx := context.Background()
	assigned := map[int]struct{}{}
	for id, _ := range s.clients {
		assigned[id] = struct{}{}
	}
	me, err := PlayerChoiceFunctions[s.Type](s.players, assigned, s)
	if err != nil {
		// TODO Mange "Room Full" error appropriately (or not)
		return err
	}
	if _, ok := s.players[me]; !ok {
		return EntityNotFound
	}
	roomData, err := s.GetState()
	if err != nil {
		return err
	}
	eventData, err := json.Marshal(event.Event{Type: "room", Payload: roomData})
	if err != nil {
		return err
	}
	if err = c.Write(ctx, websocket.MessageText, eventData); err != nil {
		return err
	}
	eventData, err = json.Marshal(event.Event{Type: "assign", Payload: []byte(fmt.Sprintf("%d", me))})
	if err != nil {
		return err
	}
	if err = c.Write(ctx, websocket.MessageText, eventData); err != nil {
		return err
	}
	transfer := make(chan *Room, 1)
	s.clients[me] = player{
		transfer: transfer,
		c:        c,
	}
	defer delete(s.clients, me)

	// TODO handle "connection closed" appropriately

	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return err
		}
		ev, err := event.ParseEvent(data)
		if err != nil {
			return err
		}
		select {
		case r := <-transfer:
			delete(s.clients, me)
			return r.Run(c)
		case s.events <- ev:
		default:
			log.Println("overload! event is not pushed", ev)
		}
	}
}

func (s *Room) ProcessEvent(e event.Event) error {
	switch e.Type {
	case "update":
		m, ok := s.elements[e.From]
		if ok {
			return m.SetState(e.Payload)
		}
		return fmt.Errorf("entity %d on %s: %w", e.From, e.Type, EntityNotFound)
	case "input":
		m, ok := s.players[e.From]
		if ok {
			return m.SetInput(e.Payload)
		}
		return fmt.Errorf("entity %d on %s: %w", e.From, e.Type, EntityNotFound)
	case "delete":
		// TODO move this in separate function
		// TODO add "send to player by id" function
		// TODO EventProcessor should get "event.Processor" with methods above
		if _, ok := s.elements[e.From]; ok {
			s.toDelete[e.From] = struct{}{}
			return nil
		} else {
			return fmt.Errorf("entity %d on %s: %w", e.From, e.Type, EntityNotFound)
		}
	case "add":
		el := elements.GenElements[e.From]()
		if err := el.SetState(e.Payload); err != nil {
			return fmt.Errorf("failed to set el state: %w", err)
		}
		s.NewElement(el)
		return nil
	case "deleted":
		s.DeleteElement(e.From)
		return nil
	// TODO add "done" event, which is closes up connections, deletes player, etc... on the other side - stop everything
	// TODO - add some kind of custom processor for "done" event so player can be tossed towards "game over room"?
	default:
		if evts, ok := EventsProcessors[s.Type]; ok {
			if f, ok := evts[e.Type]; ok {
				return f(e, s)
			}
		}
		return nil
	}
}

// TODO implement this better, add toTransfer map and transfer everyone together after update cycle, so no data races on already transferred
func (s *Room) Transfer(id int, target elements.EventProcessor) error {
	p, ok := s.clients[id]
	if !ok {
		return EntityNotFound
	}
	if target == nil {
		return errors.New("room is nil")
	}
	tg, ok := target.(*Room)
	if !ok {
		return fmt.Errorf("transfer supports only %T not %T", s, target)
	}
	select {
	case p.transfer <- tg:
		s.DeleteElement(id)
	default:
		log.Println("player is busy, cannot transfer")
	}
	return nil
}

func (s *Room) BroadcastEvent(e event.Event) {
	for _, p := range s.clients {
		data, err := json.Marshal(e)
		if err != nil {
			return
		}
		if err := p.c.Write(context.TODO(), websocket.MessageText, data); err != nil {
			return
		}
	}
}

func (s *Room) GetID() int {
	return s.ID
}

func (s *Room) GetType() string {
	return s.Type
}

func (s *Room) Draw(c canvas.Canvas) {
	for i := len(s.drawOrder) - 1; i >= 0; i-- {
		if s.drawOrder[i] == nil {
			continue
		}
		var els []elements.Drawable
		for _, el := range s.drawOrder[i] {
			els = append(els, el)
		}
		sort.Slice(els, func(i, j int) bool {
			return els[i].GetID() > els[j].GetID()
		})
		for _, e := range els {
			e.Draw(c)
		}
	}
}

func (s *Room) NewID() int {
	c := s.currentID
	s.currentID++
	return c
}

func (s *Room) NewElement(el elements.Element) {
	s.elements[el.GetID()] = el
	if p, ok := el.(elements.Playable); ok {
		s.players[el.GetID()] = p
	}
	if p, ok := el.(elements.Movable); ok {
		s.movable[el.GetID()] = p
	}
	if p, ok := el.(elements.Collidable); ok {
		s.collidable[el.GetID()] = p
	}
	if el.GetID() >= s.currentID {
		s.currentID = el.GetID() + 1
	}
	if d, ok := el.(elements.Drawable); ok {
		layer := getElementLayer(el)
		if s.drawOrder[layer] == nil {
			s.drawOrder[layer] = map[int]elements.Drawable{}
		}
		s.drawOrder[layer][el.GetID()] = d
	}
	st, err := el.GetState()
	if err != nil {
		log.Println("failed to get element's state", err)
	}
	if s.State == Running {
		s.BroadcastEvent(event.Event{Type: "add", From: el.GetType(), Payload: st})
	}
}

func getElementLayer(e elements.Element) (layer int) {
	l, ok := e.(elements.GetLayer)
	if ok {
		layer = l.GetLayer()
	}
	if layer >= layers {
		layer = layers - 1
	}
	return
}
