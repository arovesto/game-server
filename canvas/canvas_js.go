// +build js

package canvas

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	math2 "math"
	"net/http"
	"syscall/js"
	"time"

	"github.com/arovesto/gio"
	"github.com/arovesto/gio/math"
)

type WebCanvas struct {
	Window   js.Value
	Document js.Value
	Body     js.Value
	Canvas   js.Value

	ImgCtx js.Value

	Screen math.Box

	cfg gio.Config

	Images map[string]image.Image

	jsImages map[string]js.Value

	grace chan struct{}

	fps float64
}

func NewCanvas(cfg gio.Config) *WebCanvas {
	c := WebCanvas{
		Window:   js.Global(),
		cfg:      cfg,
		Images:   map[string]image.Image{},
		jsImages: map[string]js.Value{},
		grace:    make(chan struct{}),
		fps:      1000 / float64(cfg.FPSCap),
	}
	c.Document = c.Window.Get("document")
	c.Body = c.Document.Get("body")
	c.Canvas = c.Document.Call("createElement", "canvas")
	c.Canvas.Set("height", c.Window.Get("innerHeight"))
	c.Canvas.Set("width", c.Window.Get("innerWidth"))

	c.Screen.Size = math.Vector{
		X: c.Window.Get("innerWidth").Float(),
		Y: c.Window.Get("innerHeight").Float(),
	}
	c.Body.Call("appendChild", c.Canvas)
	c.ImgCtx = c.Canvas.Call("getContext", "2d")
	c.Window.Call("addEventListener", "resize", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c.Canvas.Set("height", c.Window.Get("innerHeight"))
		c.Canvas.Set("width", c.Window.Get("innerWidth"))
		c.Screen.Size = math.Vector{
			X: c.Window.Get("innerWidth").Float(),
			Y: c.Window.Get("innerHeight").Float(),
		}
		return nil
	}))
	return &c
}

func (c *WebCanvas) MoveTo(p math.Vector) {
	target := p.Sub(c.Screen.Size.Mul(0.5))
	c.Screen.Corner = c.Screen.Corner.Add((target.Sub(c.Screen.Corner)).Mul(0.2))
}

func (c *WebCanvas) Start(callback func(c *WebCanvas, duration time.Duration) bool) {
	go func() {
		var f js.Func
		var last float64
		f = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			now := args[0].Float()
			if now-last >= c.fps {
				cbkStart := time.Now()
				if callback(c, time.Duration(now-last)*time.Millisecond) {
					close(c.grace)
				}
				if float64(time.Since(cbkStart).Milliseconds()) > c.fps {
					log.Println("overload! callback is not keeping up, late on ", float64(time.Since(cbkStart).Milliseconds())-c.fps)
				}
				last = now
			}
			js.Global().Call("requestAnimationFrame", f)
			return nil
		})
		js.Global().Call("requestAnimationFrame", f)
		defer f.Release()
		<-c.grace
	}()
	<-c.grace
}

func (c *WebCanvas) Clear() {
	c.ImgCtx.Call("clearRect", 0, 0, c.Screen.Size.X, c.Screen.Size.Y)
}

type Texture struct {
	Path string
	ID   string
}

func (c *WebCanvas) DrawText(text string, where math.Vector, font string) {
	if font != "" {
		c.ImgCtx.Set("font", font)
	}
	c.ImgCtx.Call("fillText", text, where.X, where.Y)
}

func (c *WebCanvas) DrawShape(id string, world, texture math.Box) {
	if img, ok := c.jsImages[id]; !ok {
		c.jsImages[id] = js.Global().Get("Image").New()
		c.jsImages[id].Set("src", fmt.Sprintf("%s/%s", c.cfg.Server, id))
	} else {
		c.ImgCtx.Call("drawImage", img, texture.Corner.X, texture.Corner.Y, world.Size.X, world.Size.Y, world.Corner.X, world.Corner.Y, world.Size.X, world.Size.Y)
	}
}

func canvasColor(cl color.Color) (val string, opacity float64) {
	r, g, b, a := cl.RGBA()
	opacity = float64(a) / 256
	val = fmt.Sprintf("#%02X%02X%02X", r/256, g/256, b/256)
	return
}

func (c *WebCanvas) DrawColor(cl color.Color, world, texture math.Shape) {
	clr, o := canvasColor(cl)
	c.ImgCtx.Set("fillStyle", clr)
	c.ImgCtx.Set("globalAlpha", o)
	switch wrl := world.(type) {
	case math.Box:
		switch txr := texture.(type) {
		case math.Box:
			wrl.Corner = wrl.Corner.Sub(c.Screen.Corner)
			log.Println("HER")
			c.ImgCtx.Call("fillRect", wrl.Corner.X, wrl.Corner.Y, wrl.Size.X, wrl.Size.Y)
		case math.Sphere:

		default:
			log.Panicln("not implemented second type", txr)
		}
	case math.Sphere:
		switch txr := texture.(type) {
		case math.Box:
			log.Panicln("not implemented, and probably would not", wrl, txr)
		case math.Sphere:
			c.ImgCtx.Call("beginPath")
			c.ImgCtx.Call("arc", wrl.Center.X, wrl.Center.Y, wrl.R, 0, 2*math2.Pi)
			c.ImgCtx.Call("fill")
		default:
			log.Panicln("not implemented second type", txr)
		}
	default:
		log.Panicln("not implemented type", wrl)
	}
}

func (c *WebCanvas) openImage(p string) (i image.Image, e error) {
	//http.DefaultClient.Timeout = time.Second * 2
	done := make(chan struct{})
	js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			defer func() {
				close(done)
			}()
			url := fmt.Sprintf("%s/%s", c.cfg.Server, p)
			rp, err := http.Get(url)
			if err != nil {
				e = err
				return
			}
			if rp.StatusCode != http.StatusOK {
				e = fmt.Errorf("bad status: %v", rp.StatusCode)
				return
			}
			defer func() {
				_ = rp.Body.Close()
			}()
			i, e = png.Decode(rp.Body)
		}()
		return nil
	}))

	<-done
	return
}

type circle struct {
	c  math.Sphere
	cl color.Color
}

func (c circle) ColorModel() color.Model {
	return color.RGBAModel
}

func (c circle) Bounds() image.Rectangle {
	return image.Rect(int(c.c.Center.X-c.c.R), int(c.c.Center.Y-c.c.R), int(c.c.Center.X+c.c.R+1), int(c.c.Center.Y+c.c.R+1))
}

func (c circle) At(x, y int) color.Color {
	if c.c.IsInside(math.Vector{X: float64(x), Y: float64(y)}) {
		return c.cl
	} else {
		return color.RGBA{A: 0}
	}
}
