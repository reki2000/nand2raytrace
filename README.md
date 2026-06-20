# NAND-16: From Transistors to Raytracer

This learning project implements, end to end in Go, a CPU, a Forth cross-compiler,
and a full raytracer, starting from a hierarchical NAND-gate design.

## Output

![raytracer_rgb555](raytracer_rgb555.png)

The output uses quadratic sphere intersection, normal calculation,
Half-Lambert diffuse shading plus Phong specular reflection,
shadow rays, and a checkerboard ground.
Everything is computed using 8.8 fixed-point arithmetic on a 64×32 RGB555 framebuffer.

## Build and Run

```bash
go test ./...                                           # all 49 tests pass

# Assemble (MiniOS)
go run ./cmd/asmc -o boot.bin asm/boot.s                # → boot.bin (62 bytes)

# Compile Forth (raytracer)
go run ./cmd/forthc -o raytracer.bin asm/raytracer.s    # → raytracer.bin (5,782 bytes)

# Run (standalone)
go run ./cmd/nand16 raytracer.bin                       # → raytracer_rgb555.png

# Run (OS + app)
go run ./cmd/forthc -base 0x0200 -o raytracer.bin asm/raytracer.s
go run ./cmd/nand16 boot.bin raytracer.bin              # boot@0x0000 + app@0x0200
```

## Architecture Layers

### Layer 1 — Gate Simulator

`wire.go` `gate.go` `flipflop.go` `simulator.go`

An event-driven digital logic simulator built from NMOS-based NAND gates,
D flip-flops, and buses.

### Layer 2 — Combinational Logic

`logic_basic.go` `logic_arith.go` `logic_shift.go`

NOT, AND, OR, XOR, MUX, full adders, 16-bit adders/subtractors,
barrel shifters, and comparators.
All are constructed hierarchically from NAND gates.

### Layer 3 — Sequential Circuits

`sequential.go` `module.go`

16-bit registers, a register file (8×16-bit), a program counter, and a memory interface.

### Layer 4 — CPU: NAND-16

`cpu.go` `cpu_alu.go` `cpu_decode.go`

| Item | Specification |
|---|---|
| Word width | 16-bit |
| Registers | R0–R7 (R0 = zero, R6 = SP, R7 = link) |
| Instruction width | 16-bit fixed |
| Instruction formats | R/I/B/J-type |
| ALU | ADD SUB AND OR XOR SHL SHR SRA |
| Multiply | MUL(low16) / MULH(high16) |
| Memory | 64 KB byte-addressed, little-endian |
| FB | 64×32, mapped at 0xF000 (RGB555) |
| I/O | UART at 0xF800, Timer at 0xF810 |

### Layer 5 — SoC / MiniOS

`system.go` `os.go` `asm/boot.s`

SoC integration (CPU + memory + FB + UART + Timer).
The MiniOS boot code is written in assembly source [asm/boot.s](asm/boot.s) and embedded with `go:embed`.

### Layer 6 — Assembler

`assembler.go`

A two-pass assembler supporting labels, all instruction formats, and pseudo-ops.

### Layer 7 — Forth Cross-Compiler

`forth.go`

Compiles Forth source into NAND-16 machine code.

**Register convention**: R4 = TOS (cache), R6 = data stack, R5 = return stack

**Runtime**: `_udiv` (unsigned 16-bit shift-subtract division, 16 iterations)

**Word calls**: the prologue saves R7 into RSP, then the body executes, and the epilogue restores and returns with RET.
Calls beyond the 12-bit JAL range automatically fall back to register-based JALR (long call).

| Category | Words |
|---|---|
| Arithmetic | `+ - * negate abs` |
| Fixed-point | `f*` (8.8 multiply), `f/` (signed extended-precision divide), `*/` |
| Integer division | `/ mod` |
| Comparisons | `= <> < > 0= 0< 0> max min` |
| Stack | `dup drop swap over rot nip 2dup` |
| Memory | `@ ! c@ c!` |
| Control | `if else then` `begin until again` `while repeat` `do loop i j` |
| Drawing | `pixel` (8 bpp), `pixel16` (RGB555) |
| Math | `isqrt` (Newton method), `fsqrt` (fixed-point square root) |

**Constant-load optimization**: direct imm6 → 2-step ADDI → shift construction (`hi<<8+lo`) → LUI → long load.
Shift construction is preferred; even large negative values can be generated in 3–8 instructions (down from 30+ instructions in the old LUI-only approach).

**f/ sign handling**: save the dividend sign → absolute value → unsigned division → restore sign.
The previous implementation incorrectly handled negative dividends with logical right shift (SHR),
which corrupted normal calculations at x/y sign boundaries and caused spheres to split into four quadrants.

### Layer 8 — Raytracer

`cmd/nand16/main.go`

Real-time raytracing using 8.8 fixed-point arithmetic.

**Ray generation**: camera origin, pixel → screen coordinates → ray direction `(rx, ry, -256)`

**Sphere intersection**: quadratic discriminant method with overflow avoidance
```
oc = -C,  a = dot(d,d),  bh = dot(oc,d),  c = dot(oc,oc) - r²
disc = bh² - a·c,  t = (-bh - √disc) / a
```

**Normal**: `N = (P - C) / r` (accurate across all quadrants with signed `f/`)

**Lighting model**: Half-Lambert diffuse + Phong specular reflection (squared)
```
half = (dot(N,L) + 1.0) / 2      ← no terminator line
total = half × 0.625 + spec² × 0.3 + 0.1
channel = base_color × total      ← no hue shift
```
Light source: directional light `L = normalize(-1, 1, 1)` and half-vector `H = normalize(L + V)`

**Shadow rays**: intersection test from the ground hit point toward the light direction.
If the discriminant `dot(oc,L)² - (dot(oc,oc) - r²) ≥ 0` is satisfied, the point is considered in shadow and brightness is halved.

**Scene setup**:
- Warm sphere: center(80, 0, -512), r=128, base(31, 10, 4)
- Cool sphere: center(-80, -32, -384), r=96, base(6, 18, 31)
- Ground: y=-128, checkerboard (bit8 XOR approach, sign-safe)
- Background: blue gradient + warm horizon

**Forth source**: 6 defined words `isqrt` `fsqrt` `clamp0` `ground-t` `sphere-hit` `shade` `shadow?` plus the main loop

## Numerical Summary

| Item | Value |
|---|---|
| Source lines | ~3,970 |
| Test count | 49 |
| Behavioral CPU speed | ~20M instructions/sec |
| Raytracer machine code | 5,782 bytes |
| Runtime | ~120 ms |
| Resolution | 64×32 RGB555 (15 bit/pixel) |
| PNG output | 512×256 (8× upscale) |

## File Layout

```
nand16/
├── wire.go              # wires and buses
├── gate.go              # NAND gates
├── flipflop.go          # D flip-flops
├── simulator.go         # event-driven simulator
├── logic_basic.go       # NOT AND OR XOR MUX
├── logic_arith.go       # adders/subtractors/comparators
├── logic_shift.go       # barrel shifters
├── sequential.go        # registers and memory
├── module.go            # module foundation
├── cpu.go               # NAND-16 CPU
├── cpu_alu.go           # gate-level ALU
├── cpu_decode.go        # instruction decoder/encoder
├── system.go            # SoC (FB/UART/Timer)
├── os.go                # MiniOS (go:embed)
├── assembler.go         # 2-pass assembler
├── forth.go             # Forth cross-compiler
├── asm/
│   ├── boot.s           # MiniOS assembly source
│   └── raytracer.s      # raytracer Forth source
├── cmd/
│   ├── asmc/main.go     # assembler CLI (.s → .bin)
│   ├── forthc/main.go   # Forth compiler CLI (.s → .bin)
│   └── nand16/main.go   # CPU runner (.bin → PNG)
├── *_test.go ×7         # test suite (49 cases)
├── README.md
├── go.mod
└── raytracer_rgb555.png # output image
```
