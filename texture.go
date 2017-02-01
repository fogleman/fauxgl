package fauxgl

import (
	"image"
	"math"
)

type Texture interface {
	Sample(u, v float64) Color
	BilinearSample(u, v float64) Color
}

func LoadTexture(path string) (Texture, error) {
	im, err := LoadImage(path)
	if err != nil {
		return nil, err
	}
	return NewImageTexture(im), nil
}

type ImageTexture struct {
	Width  int
	Height int
	Image  image.Image
}

func NewImageTexture(im image.Image) Texture {
	size := im.Bounds().Max
	return &ImageTexture{size.X, size.Y, im}
}

func (t *ImageTexture) Sample(u, v float64) Color {
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
	return MakeColor(t.Image.At(x, y))
}

func (t *ImageTexture) BilinearSample(u, v float64) Color {
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := u * float64(t.Width)
	y := v * float64(t.Height)
	x0 := int(x)
	y0 := int(y)
	x1 := x0 + 1
	y1 := y0 + 1
	x -= float64(x0)
	y -= float64(y0)
	c00 := MakeColor(t.Image.At(x0, y0))
	c01 := MakeColor(t.Image.At(x0, y1))
	c10 := MakeColor(t.Image.At(x1, y0))
	c11 := MakeColor(t.Image.At(x1, y1))
	c := Color{}
	c = c.Add(c00.MulScalar((1 - x) * (1 - y)))
	c = c.Add(c10.MulScalar(x * (1 - y)))
	c = c.Add(c01.MulScalar((1 - x) * y))
	c = c.Add(c11.MulScalar(x * y))
	return c
}
