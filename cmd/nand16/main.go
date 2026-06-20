package main

import (
	"bytes"
	"flag"
	"fmt"
	"nand16/internal/render"
	"nand16/internal/system"
	"os"
	"time"
)

func main() {
	pngOut := flag.String("png", "raytracer_rgb555.png", "output PNG file")
	scale := flag.Int("scale", 8, "PNG scale factor")
	maxCycles := flag.Int("cycles", 500_000_000, "max CPU cycles")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: nand16 [-png out.png] app.bin [bios.bin]\n")
		fmt.Fprintf(os.Stderr, "  app.bin            loaded at 0x0000, PC starts at 0\n")
		fmt.Fprintf(os.Stderr, "  bios.bin app.bin   BIOS at 0x0000, app at 0x0200\n")
		os.Exit(1)
	}

	sys := system.NewSystem()
	var buf bytes.Buffer
	sys.Stdout = &buf

	switch flag.NArg() {
	case 1:
		// Standalone: load app at 0x0000
		app, err := os.ReadFile(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "read app: %v\n", err)
			os.Exit(1)
		}
		sys.LoadRAM(0, app)
		sys.CPU.PC = 0
		fmt.Printf("Loaded %s (%d bytes) at 0x0000\n", flag.Arg(0), len(app))

	case 2:
		// BIOS + App: BIOS at 0x0000, app at 0x0200
		bios, err := os.ReadFile(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "read bios: %v\n", err)
			os.Exit(1)
		}
		app, err := os.ReadFile(flag.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "read app: %v\n", err)
			os.Exit(1)
		}
		sys.LoadRAM(0, bios)
		sys.LoadRAM(0x0200, app)
		sys.CPU.PC = 0
		fmt.Printf("Loaded %s (%d bytes) at 0x0000, %s (%d bytes) at 0x0200\n",
			flag.Arg(0), len(bios), flag.Arg(1), len(app))
	}

	start := time.Now()
	sys.Run(*maxCycles)
	elapsed := time.Since(start)
	fmt.Printf("Executed in %v\n", elapsed)

	if *pngOut != "" {
		png := render.RGB555(sys.Mem.Data, system.FBBase, system.FBWidth, system.FBHeight, *scale)
		os.WriteFile(*pngOut, png, 0644)
		fmt.Printf("PNG: %s (%d bytes)\n", *pngOut, len(png))
	}

	if buf.Len() > 0 {
		fmt.Printf("Output: %q\n", buf.String())
	}
}
