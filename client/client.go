package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall/js"
	"time"

	"nhooyr.io/websocket"

	"github.com/arovesto/gio"
	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/event"
	"github.com/arovesto/gio/input"
	"github.com/arovesto/gio/server"
)

func RunClient(fps int, assetsPath string) {
	r := canvas.NewCanvas(gio.Config{Server: assetsPath, FPSCap: fps})
	ctx := context.Background()

	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/socket", js.Global().Get("location").Get("host").String()), nil)
	if err != nil {
		panic(fmt.Errorf("failed to create connection: %w", err))
	}
	var room server.Room
	var me elements.Playable

	inner := func() bool {
		c, cancel := context.WithTimeout(ctx, time.Millisecond)
		defer cancel()
		_, data, err := conn.Read(c)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return true
			}
			log.Println("err read", err)
			return false
		}
		e, err := event.ParseEvent(data)
		if err != nil {
			log.Println("failed parse event", err)
			return false
		}
		switch e.Type {
		case "room":
			input.ResetPressed()
			if err = room.SetState(e.Payload); err != nil {
				log.Println("failed to set room state", err)
				room = server.Room{}
			}
			room.State = server.Web
			me = nil // room changed so player is to
		case "assign":
			if room.Type == "" {
				break // room is not specified
			}
			id, err := strconv.Atoi(string(e.Payload))
			if err != nil {
				log.Println("failed to parse id for assign", err)
				return false
			}
			var ok bool
			me, ok = room.GetElement(id).(elements.Playable)
			if !ok {
				log.Println("failed to locate the player even if arrived", id)
			}
		case "game-over":
			me = nil
			log.Println("Game exited")
			// TODO think something better (maybe game should have something custom for that matter)
			os.Exit(0)
		default:
			if err = room.ProcessEvent(e); err != nil {
				log.Println("failed to process event", err)
			}
		}
		return false
	}

	js.Global().Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		for {
			if inner() {
				return nil
			}
		}
	}), 0)

	js.Global().Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if me != nil {
			i, err := me.Input()
			if err != nil {
				log.Println("failed to get player input", err)
			}
			if err := me.SetInput(i); err != nil {
				log.Println("failed to set input", err)
			}
			eventRaw, err := json.Marshal(event.Event{
				Type:    "input",
				Payload: i,
				From:    me.GetID(),
			})
			if err != nil {
				log.Println("failed to marshal event", err)
			}
			if err = conn.Write(ctx, websocket.MessageText, eventRaw); err != nil {
				log.Println("failed to send event", err)
			}
		}
		return nil
	}), 1)

	r.Start(func(c *canvas.WebCanvas, d time.Duration) (done bool) {
		c.Clear()
		if p, ok := me.(elements.PreDraw); ok {
			p.PreDraw(c)
		}
		room.Update(d)
		room.Draw(c)
		return false
	})
}
