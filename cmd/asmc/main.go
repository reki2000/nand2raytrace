package main

import (
	"flag"
	"fmt"
	"os"

	"nand16"
)

func main() {
	out := flag.String("o", "", "output .bin file (default: stdout filename based)")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: asmc [-o out.bin] input.s\n")
		os.Exit(1)
	}
	src, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
	bin, err := nand16.Assemble(string(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "assemble: %v\n", err)
		os.Exit(1)
	}
	outPath := *out
	if outPath == "" {
		outPath = flag.Arg(0) + ".bin"
	}
	if err := os.WriteFile(outPath, bin, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %d bytes\n", outPath, len(bin))
}
