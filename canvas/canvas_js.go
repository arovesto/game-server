// +build js

package canvas

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/png"
	"log"
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

	ImgCtx  js.Value
	ImgData js.Value
	ImgID   js.Value
	ImgBuff js.Value

	Frame *image.RGBA

	Screen math.Box

	cfg gio.Config

	Images map[string]image.Image

	grace chan struct{}

	fps float64
}

func NewCanvas(cfg gio.Config) *WebCanvas {
	c := WebCanvas{
		Window: js.Global(),
		cfg:    cfg,
		Images: map[string]image.Image{},
		grace:  make(chan struct{}),
		fps:    1000 / float64(cfg.FPSCap),
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
	c.ImgData = c.ImgCtx.Call("createImageData", int(c.Screen.Size.X), int(c.Screen.Size.Y))
	c.Frame = image.NewRGBA(image.Rect(0, 0, int(c.Screen.Size.X), int(c.Screen.Size.Y)))
	c.ImgBuff = js.Global().Get("Uint8Array").New(len(c.Frame.Pix))
	c.Window.Call("addEventListener", "resize", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Println("CALLED", c.Window.Get("innerWidth"), c.Canvas.Get("width"))
		c.Canvas.Set("height", c.Window.Get("innerHeight"))
		c.Canvas.Set("width", c.Window.Get("innerWidth"))
		c.Screen.Size = math.Vector{
			X: c.Window.Get("innerWidth").Float(),
			Y: c.Window.Get("innerHeight").Float(),
		}
		c.ImgData = c.ImgCtx.Call("createImageData", int(c.Screen.Size.X), int(c.Screen.Size.Y))
		c.Frame = image.NewRGBA(image.Rect(0, 0, int(c.Screen.Size.X), int(c.Screen.Size.Y)))
		c.ImgBuff = js.Global().Get("Uint8Array").New(len(c.Frame.Pix))
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
				js.CopyBytesToJS(c.ImgBuff, c.Frame.Pix)
				c.ImgData.Get("data").Call("set", c.ImgBuff)
				c.ImgCtx.Call("putImageData", c.ImgData, 0, 0)
				last = now
			}

			js.Global().Call("requestAnimationFrame", f)
			return nil
		})
		defer f.Release()
		c.ImgID = js.Global().Call("requestAnimationFrame", f)
		<-c.grace
	}()
	<-c.grace
}

func (c *WebCanvas) Stop() {
	c.Window.Call("cancelAnimationFrame", c.ImgID)
}

func (c *WebCanvas) Clear() {
	draw.Draw(c.Frame, c.Frame.Rect, image.NewUniform(image.White), image.Point{}, draw.Over)
}

type Texture struct {
	Path string
	ID   string
}

func (c *WebCanvas) DrawShape(id string, world, texture math.Shape) {
	if _, ok := c.Images[id]; !ok {
		c.Images[id] = nil
		http.DefaultClient.Timeout = time.Second * 2
		go func() {
			defer func() {
				if c.Images[id] == nil {
					delete(c.Images, id)
				}
			}()
			url := fmt.Sprintf("%s/%s", c.cfg.Server, id)
			rp, err := http.Get(url)

			if err != nil {
				log.Println("failed to http", err)
				return
			}
			if rp.StatusCode != http.StatusOK {
				log.Println("wrong status", rp.StatusCode)
				return
			}
			defer func() {
				_ = rp.Body.Close()
			}()
			var e error
			c.Images[id], e = png.Decode(rp.Body)
			if e != nil {
				log.Println("failed to decode", err)
			}
		}()
	}
	if c.Images[id] == nil {
		return
	}
	wrl := world.(math.Box)
	txr := texture.(math.Box)
	wrl.Corner = wrl.Corner.Sub(c.Screen.Corner)
	draw.Draw(c.Frame, wrl.ToImageRect(), c.Images[id], txr.Corner.ToPoint(), draw.Over)
}

func (c *WebCanvas) DrawColor(cl color.Color, world, texture math.Shape) {
	switch wrl := world.(type) {
	case math.Box:
		switch txr := texture.(type) {
		case math.Box:
			wrl.Corner = wrl.Corner.Sub(c.Screen.Corner)
			draw.Draw(c.Frame, wrl.ToImageRect(), image.NewUniform(cl), txr.Corner.ToPoint(), draw.Over)
		case math.Sphere:

		default:
			log.Panicln("not implemented second type", txr)
		}
	case math.Sphere:
		switch txr := texture.(type) {
		case math.Box:
			log.Panicln("not implemented, and probably would not", wrl, txr)
		case math.Sphere:
			circle := &circle{c: wrl, cl: cl}
			draw.Draw(c.Frame, circle.Bounds(), circle, txr.Center.Sub(math.Vector{X: txr.R, Y: txr.R}).ToPoint(), draw.Over)
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
