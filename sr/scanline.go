package sr

import (
	"image"
	"image/color"
)

type Scanline struct {
	Y, X1, X2 int
	Alpha     uint32
}

func CropScanlines(lines []Scanline, w, h int) []Scanline {
	i := 0
	for _, line := range lines {
		if line.Y < 0 || line.Y >= h {
			continue
		}
		if line.X1 >= w {
			continue
		}
		if line.X2 < 0 {
			continue
		}
		line.X1 = ClampInt(line.X1, 0, w-1)
		line.X2 = ClampInt(line.X2, 0, w-1)
		if line.X1 > line.X2 {
			continue
		}
		lines[i] = line
		i++
	}
	return lines[:i]
}

func DrawScanlines(im *image.RGBA, c color.Color, lines []Scanline) {
	const m = 0xffff
	sr, sg, sb, sa := c.RGBA()
	for _, line := range lines {
		ma := line.Alpha
		a := (m - sa*ma/m) * 0x101
		i := im.PixOffset(line.X1, line.Y)
		for x := line.X1; x <= line.X2; x++ {
			dr := uint32(im.Pix[i+0])
			dg := uint32(im.Pix[i+1])
			db := uint32(im.Pix[i+2])
			da := uint32(im.Pix[i+3])
			im.Pix[i+0] = uint8((dr*a + sr*ma) / m >> 8)
			im.Pix[i+1] = uint8((dg*a + sg*ma) / m >> 8)
			im.Pix[i+2] = uint8((db*a + sb*ma) / m >> 8)
			im.Pix[i+3] = uint8((da*a + sa*ma) / m >> 8)
			i += 4
		}
	}
}
