package gio

import (
	"fmt"
	"github.com/arovesto/gio/math"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/png"
	"net/http"
	"syscall/js"
)

type Canvas struct {
	Window js.Value
	Document js.Value
	Body js.Value
	Canvas js.Value

	ImgCtx js.Value
	ImgData js.Value
	ImgID js.Value
	ImgBuff js.Value

	Frame *image.RGBA

	Screen math.Box

	cfg Config

	Images map[string]image.Image

	grace chan struct{}


	fps float64
}

func NewCanvas(cfg Config) *Canvas {
	c := Canvas{
		Window: js.Global(),
		cfg: cfg,
		Images: map[string]image.Image{},
		grace: make(chan struct{}),
		fps: 1000 / float64(cfg.FPSCap),
	}
	c.Document = c.Window.Get("document")
	c.Body = c.Document.Get("body")
	c.Canvas = c.Document.Call("createElement", "canvas")
	c.Canvas.Set("height", c.Window.Get("innerHeight"))
	c.Canvas.Set("width", c.Window.Get("innerWidth"))

	c.Screen.Size = math.Vector{
		// TODO allow resize here
		X: c.Window.Get("innerWidth").Float(),
		Y: c.Window.Get("innerHeight").Float(),
	}
	c.Body.Call("appendChild", c.Canvas)
	c.ImgCtx = c.Canvas.Call("getContext", "2d")
	c.ImgData = c.ImgCtx.Call("createImageData", int(c.Screen.Size.X), int(c.Screen.Size.Y))
	c.Frame = image.NewRGBA(image.Rect(0, 0, int(c.Screen.Size.X), int(c.Screen.Size.Y)))
	c.ImgBuff = js.Global().Get("Uint8Array").New(len(c.Frame.Pix))
	return &c
}

func (c *Canvas) MoveTo(p math.Vector) {
	target := p.Sub(c.Screen.Size.Mul(0.5))
	c.Screen.Corner = c.Screen.Corner.Add((target.Sub(c.Screen.Corner)).Mul(0.2))
}

func (c *Canvas) Start(callback func(c *Canvas) bool) {
	go func() {
		var f js.Func
		var last float64
		f = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			now := args[0].Float()
			if now - last > c.fps { // TODO predict how long callback going to run and check it
				if callback(c) {
					close(c.grace)
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
}

func (c *Canvas) Stop() {
	c.Window.Call("cancelAnimationFrame", c.ImgID)
	<-c.grace
}

func (c *Canvas) Clear() {
	draw.Draw(c.Frame, c.Frame.Rect, image.NewUniform(image.White), image.Point{}, draw.Over)
}

type Texture struct {
	Path string
	ID string
}

func (c *Canvas) LoadTextures(textures []Texture) (err error) {
	for _, t := range textures {
		if c.Images[t.ID], err = c.openImage(t.Path); err != nil {
			return err
		}
	}
	return nil
}

func (c *Canvas) DrawShape(id string, world, texture math.Shape) {
	wrl := world.(math.Box)
	txr := texture.(math.Box)
	wrl.Corner = wrl.Corner.Sub(c.Screen.Corner)
	draw.Draw(c.Frame, wrl.ToImageRect(), c.Images[id], txr.Corner.ToPoint(), draw.Over)
}

func (c *Canvas) DrawColor(cl color.Color, world, texture math.Shape) {
	wrl := world.(math.Box)
	txr := texture.(math.Box)
	wrl.Corner = wrl.Corner.Sub(c.Screen.Corner)
	draw.Draw(c.Frame, wrl.ToImageRect(), image.NewUniform(cl), txr.Corner.ToPoint(), draw.Over)
}

func (c *Canvas) openImage(p string) (image.Image, error) {
	rp, err := http.Get(fmt.Sprintf("/assets/%s", p))
	if err != nil {
		return nil, err
	}
	if rp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %v", rp.StatusCode)
	}
	defer func() {
		_ = rp.Body.Close()
	}()
	img, err := png.Decode(rp.Body)
	return img, err
}