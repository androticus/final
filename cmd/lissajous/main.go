package main

import (
	"flag"
	"log"
)

func main() {
	run := flag.Bool("run", false, "render standalone to stdout (e.g. lissajous --run > out.gif)")
	out := flag.String("out", "lissajous.gif", "output GIF path for the embedded Final demo")
	flag.Parse()

	if *run {
		NewLissajous().Run()
		return
	}

	if err := runEmbedded(*out); err != nil {
		log.Fatalf("lissajous: %v", err)
	}
	log.Printf("lissajous: wrote %s", *out)
}
