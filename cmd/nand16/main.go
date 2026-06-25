package main

import (
	"bytes"
	"flag"
	"fmt"
	"nand16/internal/nand16"
	"nand16/internal/render"
	"nand16/internal/system"
	"os"
	"time"
)

func writePNG(sys *system.System, path string, scale int) {
	png := render.RGB555(sys.Mem.Data, system.FBBase, system.FBWidth, system.FBHeight, scale)
	os.WriteFile(path, png, 0644)
}

// stepper abstracts both behavioral and gate-level CPUs.
type stepper interface {
	step() bool
}

type behavioralStepper struct{ sys *system.System }

func (b *behavioralStepper) step() bool {
	return b.sys.CPU.Step()
}

type gateStepper struct {
	cpu *nand16.GateCPU
	mem nand16.Memory16
}

func (g *gateStepper) step() bool {
	return g.cpu.Step(g.mem)
}

func runWithSnapshots(sys *system.System, st stepper, pngOut string, every, maxCycles, scale int) int {
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
		halted := false
		for i := 0; i < n; i++ {
			if !st.step() {
				halted = true
				break
			}
		}
		remaining -= n
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
	every := flag.Int("every", 10_000, "overwrite PNG every N ticks; 0=single write at end")
	useGate := flag.Bool("gate", false, "use gate-level FlatSimulator (parallel NAND gate evaluation)")
	workers := flag.Int("workers", 0, "number of goroutine workers for -gate (0=NumCPU)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: nand16 [-gate] [-workers N] [-png out.png] app.bin [bios.bin]\n")
		fmt.Fprintf(os.Stderr, "  -gate      use the gate-level FlatSimulator instead of the behavioral CPU\n")
		fmt.Fprintf(os.Stderr, "  -workers   number of goroutine workers for gate-level parallelism (0=auto)\n")
		fmt.Fprintf(os.Stderr, "  app.bin            loaded at 0x0000, PC starts at 0\n")
		fmt.Fprintf(os.Stderr, "  bios.bin app.bin   BIOS at 0x0000, app at 0x0200\n")
		os.Exit(1)
	}

	sys := system.NewSystem()
	var buf bytes.Buffer
	sys.Stdout = &buf

	switch flag.NArg() {
	case 1:
		app, err := os.ReadFile(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "read app: %v\n", err)
			os.Exit(1)
		}
		sys.LoadRAM(0, app)
		sys.CPU.PC = 0
		fmt.Printf("Loaded %s (%d bytes) at 0x0000\n", flag.Arg(0), len(app))

	case 2:
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

	// Choose CPU engine.
	var st stepper
	if *useGate {
		gcpu := nand16.NewGateCPUParallel(*workers)
		st = &gateStepper{cpu: gcpu, mem: sys.Mem}
		fmt.Printf("Engine: gate-level FlatSimulator (%s)\n", gcpu.FlatStats())
	} else {
		st = &behavioralStepper{sys: sys}
		fmt.Println("Engine: behavioral CPU")
	}

	start := time.Now()
	snapshots := false
	if *every > 0 && *pngOut != "" {
		n := runWithSnapshots(sys, st, *pngOut, *every, *maxCycles, *scale)
		snapshots = true
		fmt.Printf("Wrote %d PNG snapshots to %s (every %d ticks)\n", n, *pngOut, *every)
	} else {
		for i := 0; i < *maxCycles; i++ {
			if !st.step() {
				break
			}
		}
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
