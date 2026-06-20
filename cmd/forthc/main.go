package main

import (
	"flag"
	"fmt"
	"os"

	"nand16"
)

func main() {
	out := flag.String("o", "", "output .bin file")
	base := flag.Int("base", 0, "base address for code generation (e.g. 0x0200)")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: forthc [-o out.bin] [-base 0x0200] input.s\n")
		os.Exit(1)
	}
	src, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
	fc := nand16.NewForthCompiler()
	bin, err := fc.Compile(string(src), *base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compile: %v\n", err)
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
	fmt.Printf("%s: %d bytes (base=0x%04X)\n", outPath, len(bin), *base)
}
