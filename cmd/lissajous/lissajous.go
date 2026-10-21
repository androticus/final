package main

import (
	"image"
	"image/color"
	"image/gif"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
)

var palette = []color.Color{color.Black, color.RGBA{0, 255, 0, 255}}

const (
	backIndex = 0 // first color in palette
	foreIndex = 1 // next color in palette
)

// Lissajous generates an animated GIF of random Lissajous figures.
type Lissajous struct {
	cycles  int     // number of complete x oscillator revolutions
	res     float64 // angular resolution
	size    int     // image canvas covers [-size..+size]
	nframes int     // number of animation frames
	delay   int     // delay between frames in 10ms units
}

func NewLissajous() *Lissajous {
	return &Lissajous{
		cycles:  5,
		res:     0.001,
		size:    100,
		nframes: 64,
		delay:   8,
	}
}

// Run renders the animated Lissajous GIF to stdout; pipe it to a file
// (e.g. `lissajous --run > out.gif`).
func (l *Lissajous) Run() {
	if err := l.Generate(os.Stdout); err != nil {
		log.Fatalf("lissajous: encoding GIF: %v", err)
	}
}

// Generate renders the animated Lissajous GIF to out.
//
// Adapted from Donovan & Kernighan, "The Go Programming Language", chapter 1.
func (l *Lissajous) Generate(out io.Writer) error {
	freq := rand.Float64() * 3.0 // relative frequency of y oscillator
	anim := gif.GIF{LoopCount: l.nframes}
	phase := 0.0 // phase difference
	for i := 0; i < l.nframes; i++ {
		rect := image.Rect(0, 0, 2*l.size+1, 2*l.size+1)
		img := image.NewPaletted(rect, palette)
		for t := 0.0; t < float64(l.cycles)*2*math.Pi; t += l.res {
			x := math.Sin(t)
			y := math.Sin(t*freq + phase)
			img.SetColorIndex(
				l.size+int(x*float64(l.size)+0.5),
				l.size+int(y*float64(l.size)+0.5),
				foreIndex,
			)
		}
		phase += 0.1
		anim.Delay = append(anim.Delay, l.delay)
		anim.Image = append(anim.Image, img)
	}
	return gif.EncodeAll(out, &anim)
}
