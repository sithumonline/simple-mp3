package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/basicwidget/cjkfont"
	"golang.org/x/text/language"

	"github.com/sithumonline/simple-mp3/pkg"
	"github.com/sithumonline/simple-mp3/pkg/art"
)

type modelKey int

const (
	modelKeyModel modelKey = iota
)

type Root struct {
	guigui.DefaultWidget

	background        basicwidget.Background
	playOrPauseButton basicwidget.Button
	nextButton        basicwidget.Button
	prevButton        basicwidget.Button
	trackLabelText    basicwidget.Text
	volUpButton       basicwidget.Button
	volDownButton     basicwidget.Button

	art *art.GenArt

	model *pkg.Player

	locales           []language.Tag
	faceSourceEntries []basicwidget.FaceSourceEntry

	fontsReady bool
	wired      bool
	lastLabel  string
}

func (r *Root) updateFontFaceSources(ctx *guigui.Context) {
	if r.fontsReady {
		return
	} // <-- guard

	// 1) locales
	r.locales = r.locales[:0]
	r.locales = ctx.AppendLocales(r.locales)

	// 2) (Re)build face list from scratch
	r.faceSourceEntries = r.faceSourceEntries[:0]

	// --- Noto Sans Symbols 2 ---
	data, err := os.ReadFile("NotoSansSymbols2-Regular.ttf")
	if err != nil {
		log.Printf("read font failed: %v", err)
	} else if src, err := text.NewGoTextFaceSource(bytes.NewReader(data)); err != nil {
		log.Printf("parse font failed: %v", err)
	} else {
		// Put it FIRST so it handles symbols and simple Latin.
		r.faceSourceEntries = append(r.faceSourceEntries, basicwidget.FaceSourceEntry{
			FaceSource: src,
			UnicodeRanges: []basicwidget.UnicodeRange{
				// Basic Latin (space + ASCII letters, just in case)
				{Min: 0x0020, Max: 0x007E},
				// General Punctuation (safer spacing/punct)
				{Min: 0x2000, Max: 0x206F},
				// Arrows (extra icons you might use)
				{Min: 0x2190, Max: 0x21FF},
				// Misc Technical (⏮ ⏯ ⏭ ⏸ live here: U+23EE..U+23F8)
				{Min: 0x2300, Max: 0x23FF},
				// Geometric Shapes (▶ = U+25B6)
				{Min: 0x25A0, Max: 0x25FF},
				// Misc Symbols & Arrows (extra triangles/controls)
				{Min: 0x2B00, Max: 0x2BFF},
			},
		})
	}

	// --- CJK fallbacks after symbols so they don't "claim" the ranges ---
	r.faceSourceEntries = cjkfont.AppendRecommendedFaceSourceEntries(
		r.faceSourceEntries, r.locales,
	)

	// 3) Register
	basicwidget.SetFaceSources(r.faceSourceEntries)

	r.fontsReady = true
}

func (r *Root) Model(key any) any {
	switch key {
	case modelKeyModel:
		return r.model
	default:
		return nil
	}
}

func (r *Root) AddChildren(context *guigui.Context, adder *guigui.ChildAdder) {
	adder.AddChild(&r.background)
	adder.AddChild(&r.playOrPauseButton)
	if r.art != nil {
		adder.AddChild(r.art)
	}
	adder.AddChild(&r.nextButton)
	adder.AddChild(&r.prevButton)
	adder.AddChild(&r.trackLabelText)
	adder.AddChild(&r.volUpButton)
	adder.AddChild(&r.volDownButton)
}

func (r *Root) Update(context *guigui.Context) error {
	r.updateFontFaceSources(context)

	model := r.model
	//if !r.wired {
	r.prevButton.SetOnUp(func() { _ = model.Prev() })
	r.playOrPauseButton.SetOnUp(func() { model.PauseToggle() })
	r.nextButton.SetOnUp(func() { _ = model.Next() })
	r.volUpButton.SetOnUp(func() { model.VolUp(0.5) })
	r.volDownButton.SetOnUp(func() { model.VolDown(0.5) })
	r.wired = true

	//context.SetOnKeyJustPressed(func(key ebiten.Key) bool {
	//	switch key {
	//	case ebiten.KeySpace:
	//		model.PauseToggle()
	//		return true
	//	case ebiten.KeyLeft:
	//		_ = model.Prev()
	//		return true
	//	case ebiten.KeyRight:
	//		_ = model.Next()
	//		return true
	//	}
	//	return false
	//})
	//}

	if model.Cur != nil && !model.Cur.Ctrl.Paused {
		r.playOrPauseButton.SetText(pkg.TextOnly("⏸"))
	} else {
		r.playOrPauseButton.SetText(pkg.TextOnly("▶"))
	}
	r.prevButton.SetText(pkg.TextOnly("⏮"))
	r.nextButton.SetText(pkg.TextOnly("⏭"))

	r.volUpButton.SetText(pkg.TextOnly("+"))
	r.volDownButton.SetText(pkg.TextOnly("-"))

	var labelTxt string
	switch {
	case model.Cur == nil:
		labelTxt = "No track"
	case model.Cur.Ctrl.Paused:
		labelTxt = "Paused: " + pkg.EllipsizeMiddle(filepath.Base(model.Cur.Path), 36)
	default:
		labelTxt = "Playing: " + pkg.EllipsizeMiddle(filepath.Base(model.Cur.Path), 36)
	}

	if labelTxt != r.lastLabel {
		r.trackLabelText.SetValue(labelTxt)
		r.lastLabel = labelTxt
	}

	return nil
}

func (r *Root) Layout(context *guigui.Context, widget guigui.Widget) image.Rectangle {
	switch widget {
	case &r.background:
		return context.Bounds(r)
	}

	u := basicwidget.UnitSize(context)
	return (guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items: []guigui.LinearLayoutItem{
			{
				Size: guigui.FixedSize(u),
				Layout: guigui.LinearLayout{
					Direction: guigui.LayoutDirectionHorizontal,
					Items: []guigui.LinearLayoutItem{
						{
							Widget: &r.trackLabelText,
							Size:   guigui.FlexibleSize(1),
						},
					},
					Gap: u / 2,
				},
			},
			{
				Widget: r.art,
				Size:   guigui.FlexibleSize(1),
			},
			{
				Size: guigui.FixedSize(u),
				Layout: guigui.LinearLayout{
					Direction: guigui.LayoutDirectionHorizontal,
					Items: []guigui.LinearLayoutItem{
						{
							Widget: &r.volDownButton,
							Size:   guigui.FixedSize(u),
						},
						{
							Widget: &r.prevButton,
							Size:   guigui.FixedSize(3 * u),
						},
						{
							Widget: &r.playOrPauseButton,
							Size:   guigui.FixedSize(3 * u),
						},
						{
							Widget: &r.nextButton,
							Size:   guigui.FixedSize(3 * u),
						},
						{
							Widget: &r.volUpButton,
							Size:   guigui.FixedSize(u),
						},
					},
					Gap: u / 2,
				},
			},
		},
		Gap: u / 2,
	}).WidgetBounds(context, context.Bounds(r).Inset(u/2), widget)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mp3player <file1.mp3> [file2.mp3 ...]")
		os.Exit(1)
	}
	// Basic existence check
	var files []string
	for _, a := range os.Args[1:] {
		if _, err := os.Stat(a); err != nil {
			log.Fatalf("cannot open %s: %v", a, err)
		}
		files = append(files, a)
	}

	player := pkg.NewPlayer(files)
	if err := player.Play(0); err != nil {
		log.Fatal(err)
	}
	go player.RunAutoAdvance()

	op := &guigui.RunOptions{
		Title:         "Music Player",
		WindowMinSize: image.Pt(350, 300),
		WindowMaxSize: image.Pt(350, 300),
		RunGameOptions: &ebiten.RunGameOptions{
			ApplePressAndHoldEnabled: true,
		},
	}
	if err := guigui.Run(&Root{
		model: player,
		art:   art.NewGenArt(player, 640, 260*1.5),
	}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
