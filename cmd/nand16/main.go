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

// writePNG snapshots the current framebuffer to a PNG file.
func writePNG(sys *system.System, path string, scale int) {
	png := render.RGB555(sys.Mem.Data, system.FBBase, system.FBWidth, system.FBHeight, scale)
	os.WriteFile(path, png, 0644)
}

// runWithSnapshots steps the CPU, overwriting the PNG file every `every` cycles
// so the latest framebuffer can be watched in an auto-reloading image viewer. It
// stops when the CPU halts or the cycle budget is exhausted, and returns the
// number of snapshots written.
func runWithSnapshots(sys *system.System, pngOut string, every, maxCycles, scale int) int {
	if every < 1 {
		every = 1
	}
	shots := 0
	remaining := maxCycles
	for remaining > 0 {
		n := every
		if n > remaining {
			n = remaining
		}
		ran, halted := sys.StepN(n)
		remaining -= ran
		writePNG(sys, pngOut, scale)
		shots++
		if halted {
			break
		}
	}
	return shots
}

func main() {
	pngOut := flag.String("png", "raytracer_rgb555.png", "output PNG file")
	scale := flag.Int("scale", 8, "PNG/window scale factor")
	maxCycles := flag.Int("cycles", 500_000_000, "max CPU cycles")
	every := flag.Int("every", 10_000, "for PNG output: overwrite the PNG with the current framebuffer every "+
		"N CPU ticks; 0 disables periodic snapshots and writes a single PNG at the end")
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
	snapshots := false
	if *every > 0 && *pngOut != "" {
		n := runWithSnapshots(sys, *pngOut, *every, *maxCycles, *scale)
		snapshots = true
		fmt.Printf("Wrote %d PNG snapshots to %s (every %d ticks)\n", n, *pngOut, *every)
	} else {
		sys.Run(*maxCycles)
	}
	elapsed := time.Since(start)
	fmt.Printf("Executed in %v\n", elapsed)

	if *pngOut != "" && !snapshots {
		png := render.RGB555(sys.Mem.Data, system.FBBase, system.FBWidth, system.FBHeight, *scale)
		os.WriteFile(*pngOut, png, 0644)
		fmt.Printf("PNG: %s (%d bytes)\n", *pngOut, len(png))
	}

	if buf.Len() > 0 {
		fmt.Printf("Output: %q\n", buf.String())
	}
}
