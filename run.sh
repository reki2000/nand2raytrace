#!/usr/bin/env bash
# Build and run the full NAND-16 toolchain to produce the raytracer PNG.
#
#   bios.s ──asmc──────────────► bios.bin   (loaded at 0x0000)
#   raytracer.fth ──forthc──► .s ──asmc -base 0x0200──► raytracer.bin (loaded at 0x0200)
#   nand16 bios.bin raytracer.bin ─────────────────────► raytracer_rgb555.png
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p bin

PNG="${1:-raytracer_rgb555.png}"

echo "[1/4] Assembling BIOS..."
go run ./cmd/asmc -o bin/bios.bin asm/bios.s

echo "[2/4] Compiling Forth raytracer to assembly..."
go run ./cmd/forthc -o bin/raytracer.s asm/raytracer.fth

echo "[3/4] Assembling raytracer for load address 0x0200..."
go run ./cmd/asmc -base 0x0200 -o bin/raytracer.bin bin/raytracer.s

echo "[4/4] Running on the NAND-16 simulator (BIOS @ 0x0000 + app @ 0x0200)..."
go run ./cmd/nand16 -png "$PNG" bin/bios.bin bin/raytracer.bin

echo "Wrote $PNG"
