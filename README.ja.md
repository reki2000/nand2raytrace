# NAND-16: From Transistors to Raytracer

NANDゲートの階層的構成からCPU、Forthクロスコンパイラ、
本格レイトレーサまでをGoで一気通貫に実装する学習プロジェクト。

## 出力

![raytracer_rgb555](raytracer_rgb555.png)

二次方程式による球面交差、法線計算、Half-Lambert拡散＋Phong鏡面反射、
シャドウレイ、チェッカーボード地面。
すべて8.8固定小数点演算、64×32 RGB555。

## ビルド・実行

```bash
go test ./...                                           # 49テスト全PASS

# アセンブル (MiniOS)
go run ./cmd/asmc -o boot.bin asm/boot.s                # → boot.bin (62 bytes)

# Forthコンパイル (レイトレーサ)
go run ./cmd/forthc -o raytracer.bin asm/raytracer.s    # → raytracer.bin (5,782 bytes)

# 実行 (スタンドアロン)
go run ./cmd/nand16 raytracer.bin                       # → raytracer_rgb555.png

# 実行 (OS + アプリ)
go run ./cmd/forthc -base 0x0200 -o raytracer.bin asm/raytracer.s
go run ./cmd/nand16 boot.bin raytracer.bin              # boot@0x0000 + app@0x0200
```

## アーキテクチャ全層

### Layer 1 — ゲートシミュレータ

`wire.go` `gate.go` `flipflop.go` `simulator.go`

イベント駆動型デジタル論理シミュレータ。NMOSベースNANDゲート、Dフリップフロップ、バス。

### Layer 2 — 組み合わせ論理

`logic_basic.go` `logic_arith.go` `logic_shift.go`

NOT, AND, OR, XOR, MUX, 全加算器, 16bit加減算器, バレルシフタ, 比較器。
すべてNANDゲートから階層的に構成。

### Layer 3 — 順序回路

`sequential.go` `module.go`

16bitレジスタ、レジスタファイル（8×16bit）、プログラムカウンタ、メモリインターフェース。

### Layer 4 — CPU「NAND-16」

`cpu.go` `cpu_alu.go` `cpu_decode.go`

| 項目 | 仕様 |
|---|---|
| ワード幅 | 16bit |
| レジスタ | R0–R7（R0=ゼロ, R6=SP, R7=リンク） |
| 命令幅 | 16bit固定 |
| 命令形式 | R/I/B/J-type |
| ALU | ADD SUB AND OR XOR SHL SHR SRA |
| 乗算 | MUL(low16) / MULH(high16) |
| メモリ | 64KB バイトアドレス, リトルエンディアン |
| FB | 64×32, メモリマップ 0xF000 (RGB555) |
| I/O | UART 0xF800, Timer 0xF810 |

### Layer 5 — SoC / MiniOS

`system.go` `os.go` `asm/boot.s`

SoC統合（CPU＋メモリ＋FB＋UART＋Timer）。
MiniOSブートコードはアセンブリソース `asm/boot.s` で記述し `go:embed` で組み込み。

### Layer 6 — アセンブラ

`assembler.go`

2パスアセンブラ。ラベル、全命令形式、疑似命令。

### Layer 7 — Forthクロスコンパイラ

`forth.go`

Forthソースを NAND-16 機械語にコンパイル。

**レジスタ規約**: R4=TOS(キャッシュ), R6=データスタック, R5=リターンスタック

**ランタイム**: `_udiv` (符号なし16bit shift-subtract除算, 16反復)

**ワード呼び出し**: プロローグでR7をRSPに退避 → 本体 → エピローグで復帰・RET。
12bit JAL範囲を超える呼び出しはレジスタ経由JALR(ロングコール)に自動フォールバック。

| カテゴリ | ワード |
|---|---|
| 算術 | `+ - * negate abs` |
| 固定小数点 | `f*` (8.8乗算), `f/` (符号付き拡張精度除算), `*/` |
| 整数除算 | `/ mod` |
| 比較 | `= <> < > 0= 0< 0> max min` |
| スタック | `dup drop swap over rot nip 2dup` |
| メモリ | `@ ! c@ c!` |
| 制御 | `if else then` `begin until again` `while repeat` `do loop i j` |
| 描画 | `pixel` (8bpp), `pixel16` (RGB555) |
| 数学 | `isqrt` (Newton法), `fsqrt` (固定小数点平方根) |

**定数ロード最適化**: imm6直接 → 2段ADDI → シフト構築(`hi<<8+lo`) → LUI → ロングロード。
シフト構築を優先し、負の大きな値でも3〜8命令で生成（旧LUI方式の30+命令から大幅削減）。

**f/ 符号対応**: 被除数の符号を保存→絶対値化→unsigned除算→符号復帰。
旧実装は論理右シフト(SHR)で負の被除数を誤処理し、
法線計算がx/y軸の符号境界で壊れて球面が4象限に分割される致命バグがあった。

### Layer 8 — レイトレーサ

`cmd/main.go`

8.8固定小数点によるリアルタイムレイトレーシング。

**レイ生成**: カメラ原点、ピクセル→スクリーン座標→レイ方向 `(rx, ry, -256)`

**球面交差**: 二次方程式の半判別式方式でオーバーフロー回避
```
oc = -C,  a = dot(d,d),  bh = dot(oc,d),  c = dot(oc,oc) - r²
disc = bh² - a·c,  t = (-bh - √disc) / a
```

**法線**: `N = (P - C) / r` （符号付きf/で全象限正確）

**照明モデル**: Half-Lambert拡散 + Phong鏡面反射(²乗)
```
half = (dot(N,L) + 1.0) / 2      ← ターミネーター線なし
total = half × 0.625 + spec² × 0.3 + 0.1
channel = base_color × total      ← 色相シフトなし
```
光源: 平行光 L = normalize(-1, 1, 1)、半ベクトル H = normalize(L + V)

**シャドウレイ**: 地面ヒットポイントから光源方向への交差判定。
判別式 `dot(oc,L)² - (dot(oc,oc) - r²) ≥ 0` で影の中なら輝度半減。

**シーン構成**:
- 暖色球: center(80, 0, -512), r=128, base(31, 10, 4)
- 寒色球: center(-80, -32, -384), r=96, base(6, 18, 31)
- 地面: y=-128, チェッカーボード（bit8 XOR方式, 符号安全）
- 背景: 青グラデーション + 地平線暖色

**Forthソース**: 6ワード定義 `isqrt` `fsqrt` `clamp0` `ground-t` `sphere-hit` `shade` `shadow?` + メインループ

## 数値

| 項目 | 値 |
|---|---|
| ソース行数 | ~3,970 |
| テスト数 | 49 |
| behavioral CPU速度 | ~20M命令/秒 |
| レイトレーサ機械語 | 5,782 bytes |
| 実行時間 | ~120ms |
| 解像度 | 64×32 RGB555 (15bit/pixel) |
| PNG出力 | 512×256 (8倍拡大) |

## ファイル構成

```
nand16/
├── wire.go              # ワイヤ・バス
├── gate.go              # NANDゲート
├── flipflop.go          # Dフリップフロップ
├── simulator.go         # イベント駆動シミュレータ
├── logic_basic.go       # NOT AND OR XOR MUX
├── logic_arith.go       # 加算器 減算器 比較器
├── logic_shift.go       # バレルシフタ
├── sequential.go        # レジスタ メモリ
├── module.go            # モジュール基盤
├── cpu.go               # NAND-16 CPU
├── cpu_alu.go           # ゲートレベルALU
├── cpu_decode.go        # 命令デコーダ・エンコーダ
├── system.go            # SoC (FB/UART/Timer)
├── os.go                # MiniOS (go:embed)
├── assembler.go         # 2パスアセンブラ
├── forth.go             # Forthクロスコンパイラ
├── asm/
│   ├── boot.s           # MiniOSアセンブリソース
│   └── raytracer.s      # レイトレーサForthソース
├── cmd/
│   ├── asmc/main.go     # アセンブラCLI (.s → .bin)
│   ├── forthc/main.go   # ForthコンパイラCLI (.s → .bin)
│   └── nand16/main.go   # CPUランナー (.bin → PNG)
├── *_test.go ×7         # テスト群 (49件)
├── README.md
├── go.mod
└── raytracer_rgb555.png # 出力画像
```
