package main

import (
	"flag"
	"fmt"
	"nand16/internal/forthc"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	out := flag.String("o", "", "output .s file (default: input + .s)")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: forthc [-o out.s] input.fth\n")
		os.Exit(1)
	}
	src, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
	fc := forthc.NewForthCompiler()
	asm, err := fc.Compile(string(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "compile: %v\n", err)
		os.Exit(1)
	}
	outPath := *out
	if outPath == "" {
		in := flag.Arg(0)
		outPath = strings.TrimSuffix(in, filepath.Ext(in)) + ".s"
	}
	if err := os.WriteFile(outPath, []byte(asm), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %d bytes\n", outPath, len(asm))
}
