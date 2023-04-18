package fauxgl

import (
	"image"
)

type ImageBuffer struct {
	Height     int
	Width      int
	Buffer     *image.NRGBA
	AlphaBlend bool
}

func NewImageBuffer(width, height int) *ImageBuffer {
	b := &ImageBuffer{}

	b.Width = width
	b.Height = height
	b.Buffer = image.NewNRGBA(image.Rect(0, 0, width, height))

	return b
}

func (b *ImageBuffer) Clear(color Color) {
	c := color.NRGBA()
	for y := 0; y < b.Height; y++ {
		i := b.Buffer.PixOffset(0, y)
		for x := 0; x < b.Width; x++ {
			b.Buffer.Pix[i+0] = c.R
			b.Buffer.Pix[i+1] = c.G
			b.Buffer.Pix[i+2] = c.B
			b.Buffer.Pix[i+3] = c.A
			i += 4
		}
	}
}

func (b *ImageBuffer) Write(x, y int, color Color) {
	// update color buffer
	if b.AlphaBlend && color.A < 1 {
		sr, sg, sb, sa := color.NRGBA().RGBA()
		a := (0xffff - sa) * 0x101
		j := b.Buffer.PixOffset(x, y)
		dr := &b.Buffer.Pix[j+0]
		dg := &b.Buffer.Pix[j+1]
		db := &b.Buffer.Pix[j+2]
		da := &b.Buffer.Pix[j+3]
		*dr = uint8((uint32(*dr)*a/0xffff + sr) >> 8)
		*dg = uint8((uint32(*dg)*a/0xffff + sg) >> 8)
		*db = uint8((uint32(*db)*a/0xffff + sb) >> 8)
		*da = uint8((uint32(*da)*a/0xffff + sa) >> 8)
	} else {
		b.Buffer.SetNRGBA(x, y, color.NRGBA())
	}
}

func (b *ImageBuffer) Dimensions() (int, int) {
	return b.Width, b.Height
}

func (b *ImageBuffer) ToImage() image.Image {
	return b.Buffer
}
