package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"time"

	"nand16"
)

func main() {
	pngOut := flag.String("png", "raytracer_rgb555.png", "output PNG file")
	scale := flag.Int("scale", 8, "PNG scale factor")
	maxCycles := flag.Int("cycles", 500_000_000, "max CPU cycles")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: nand16 [-png out.png] app.bin [boot.bin]\n")
		fmt.Fprintf(os.Stderr, "  app.bin            loaded at 0x0000, PC starts at 0\n")
		fmt.Fprintf(os.Stderr, "  boot.bin app.bin   boot at 0x0000, app at 0x0200\n")
		os.Exit(1)
	}

	sys := nand16.NewSystem()
	var buf bytes.Buffer
	sys.Stdout = &buf

	switch flag.NArg() {
	case 1:
		// Standalone: load app at 0x0000
		app, err := os.ReadFile(flag.Arg(0))
		if err != nil { fmt.Fprintf(os.Stderr, "read app: %v\n", err); os.Exit(1) }
		sys.LoadRAM(0, app)
		sys.CPU.PC = 0
		fmt.Printf("Loaded %s (%d bytes) at 0x0000\n", flag.Arg(0), len(app))

	case 2:
		// Boot + App: boot at 0x0000, app at 0x0200
		boot, err := os.ReadFile(flag.Arg(0))
		if err != nil { fmt.Fprintf(os.Stderr, "read boot: %v\n", err); os.Exit(1) }
		app, err := os.ReadFile(flag.Arg(1))
		if err != nil { fmt.Fprintf(os.Stderr, "read app: %v\n", err); os.Exit(1) }
		sys.LoadRAM(0, boot)
		sys.LoadRAM(0x0200, app)
		sys.CPU.PC = 0
		fmt.Printf("Loaded %s (%d bytes) at 0x0000, %s (%d bytes) at 0x0200\n",
			flag.Arg(0), len(boot), flag.Arg(1), len(app))
	}

	start := time.Now()
	sys.Run(*maxCycles)
	elapsed := time.Since(start)
	fmt.Printf("Executed in %v\n", elapsed)

	if *pngOut != "" {
		png := sys.RenderPNGRGB555(*scale)
		os.WriteFile(*pngOut, png, 0644)
		fmt.Printf("PNG: %s (%d bytes)\n", *pngOut, len(png))
	}

	if buf.Len() > 0 {
		fmt.Printf("Output: %q\n", buf.String())
	}
}
