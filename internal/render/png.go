// Package render encodes framebuffer/memory contents into image formats.
// It is decoupled from the machine core: callers pass raw bytes and
// geometry, keeping image/png out of the system package.
package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// pngImage implements image.Image over a packed RGBA byte slice.
type pngImage struct {
	w, h int
	rgba []byte
}

func (p *pngImage) ColorModel() color.Model { return color.RGBAModel }
func (p *pngImage) Bounds() image.Rectangle { return image.Rect(0, 0, p.w, p.h) }
func (p *pngImage) At(x, y int) color.Color {
	off := (y*p.w + x) * 4
	return color.RGBA{p.rgba[off], p.rgba[off+1], p.rgba[off+2], p.rgba[off+3]}
}

// encodePNG encodes a packed RGBA buffer of size w*h*4 to PNG bytes.
func encodePNG(w, h int, rgba []byte) []byte {
	var buf bytes.Buffer
	png.Encode(&buf, &pngImage{w: w, h: h, rgba: rgba})
	return buf.Bytes()
}

// RGB555 reads a framebuffer of 16-bit RGB555 pixels from mem starting at base
// and encodes it to a PNG, scaled by an integer factor.
// Each pixel is 2 bytes (little-endian): 0RRRRRGGGGGBBBBB.
func RGB555(mem []byte, base uint16, fbWidth, fbHeight, scale int) []byte {
	if scale < 1 {
		scale = 1
	}
	w := fbWidth * scale
	h := fbHeight * scale

	rgba := make([]byte, w*h*4)
	for y := 0; y < fbHeight; y++ {
		for x := 0; x < fbWidth; x++ {
			addr := base + uint16((y*fbWidth+x)*2)
			pixel := uint16(mem[addr]) | uint16(mem[addr+1])<<8

			r5 := (pixel >> 10) & 0x1F
			g5 := (pixel >> 5) & 0x1F
			b5 := pixel & 0x1F

			r8 := byte(r5 * 255 / 31)
			g8 := byte(g5 * 255 / 31)
			b8 := byte(b5 * 255 / 31)

			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					off := ((y*scale+sy)*w + x*scale + sx) * 4
					rgba[off+0] = r8
					rgba[off+1] = g8
					rgba[off+2] = b8
					rgba[off+3] = 255
				}
			}
		}
	}
	return encodePNG(w, h, rgba)
}
