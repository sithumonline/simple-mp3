// ga_art.go (in your main package)
package art

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"

	"github.com/sithumonline/simple-mp3/pkg"
)

type GenArt struct {
	guigui.DefaultWidget

	model    *pkg.Player // so we can read CurrentLevel()
	c        *generativeart.Canva
	tex      *ebiten.Image
	last     time.Time
	interval time.Duration // redraw cadence
	start    time.Time
}

func NewGenArt(model *pkg.Player, w, h int) *GenArt {
	g := &GenArt{
		model:    model,
		c:        generativeart.NewCanva(w, h),
		interval: 70 * time.Millisecond, // ~14 fps
		start:    time.Now(),
	}
	g.c.SetBackground(color.RGBA{8, 10, 20, 255})
	g.c.FillBackground()
	return g
}

func (g *GenArt) Update(*guigui.Context) error {
	// throttle expensive regeneration
	if time.Since(g.last) < g.interval {
		return nil
	}
	g.last = time.Now()

	// audio level 0..1 (use your LevelMeter-backed accessor)
	lv := 0.0
	if g.model != nil {
		lv = g.model.CurrentLevel()
	}

	// Animate parameters with time + level
	t := time.Since(g.start).Seconds()
	depth := 3 + int(6.0*lv+3.0*math.Sin(t*0.8))
	if depth < 1 {
		depth = 1
	}

	// IMPORTANT: generativeart arts typically use math/rand global.
	// Seed the *global* generator so it actually affects the art.
	// Mix time and level so frames differ.
	seed := time.Now().UnixNano() ^ int64(lv*1e6)
	rand.Seed(seed)

	// re-render a fresh frame
	g.c.SetBackground(color.RGBA{8, 10, 20, 255})
	g.c.FillBackground()
	g.c.Draw(arts.NewCircleLoop2(depth)) // try other arts.* too

	// upload raw RGBA to ebiten texture
	img := g.c.Img() // *image.RGBA
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if g.tex == nil || g.tex.Bounds().Dx() != w || g.tex.Bounds().Dy() != h {
		g.tex = ebiten.NewImage(w, h)
	}
	if img.Stride == 4*w {
		g.tex.WritePixels(img.Pix) // fast path
	} else {
		buf := make([]byte, 4*w*h)
		for y := 0; y < h; y++ {
			copy(buf[y*4*w:(y+1)*4*w], img.Pix[y*img.Stride:y*img.Stride+4*w])
		}
		g.tex.WritePixels(buf)
	}
	return nil
}

func (g *GenArt) Layout(ctx *guigui.Context, _ guigui.Widget) image.Rectangle { return ctx.Bounds(g) }

func (g *GenArt) Draw(ctx *guigui.Context, dst *ebiten.Image) {
	if g.tex == nil {
		return
	}
	b := ctx.Bounds(g)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.Min.X), float64(b.Min.Y))
	dst.DrawImage(g.tex, op)
}
