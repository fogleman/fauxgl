package fauxgl

import (
	"fmt"
	"image/color"
	"math"
	"strings"
)

var (
	Discard     = Color{}
	Transparent = Color{}
	Black       = Color{0, 0, 0, 1}
	White       = Color{1, 1, 1, 1}
)

type Color struct {
	R, G, B, A float64
}

func Gray(x float64) Color {
	return Color{x, x, x, 1}
}

func MakeColor(c color.Color) Color {
	r, g, b, a := c.RGBA()
	const d = 0xffff
	return Color{float64(r) / d, float64(g) / d, float64(b) / d, float64(a) / d}
}

func HexColor(x string) Color {
	x = strings.Trim(x, "#")
	var r, g, b, a int
	a = 255
	switch len(x) {
	case 3:
		fmt.Sscanf(x, "%1x%1x%1x", &r, &g, &b)
		r = (r << 4) | r
		g = (g << 4) | g
		b = (b << 4) | b
	case 4:
		fmt.Sscanf(x, "%1x%1x%1x%1x", &r, &g, &b, &a)
		r = (r << 4) | r
		g = (g << 4) | g
		b = (b << 4) | b
		a = (a << 4) | a
	case 6:
		fmt.Sscanf(x, "%02x%02x%02x", &r, &g, &b)
	case 8:
		fmt.Sscanf(x, "%02x%02x%02x%02x", &r, &g, &b, &a)
	}
	const d = 0xff
	return Color{float64(r) / d, float64(g) / d, float64(b) / d, float64(a) / d}
}

func (c Color) NRGBA() color.NRGBA {
	const d = 0xff
	r := Clamp(c.R, 0, 1)
	g := Clamp(c.G, 0, 1)
	b := Clamp(c.B, 0, 1)
	a := Clamp(c.A, 0, 1)
	return color.NRGBA{uint8(r * d), uint8(g * d), uint8(b * d), uint8(a * d)}
}

func (a Color) Opaque() Color {
	return Color{a.R, a.G, a.B, 1}
}

func (a Color) Alpha(alpha float64) Color {
	return Color{a.R, a.G, a.B, alpha}
}

func (a Color) Lerp(b Color, t float64) Color {
	return a.Add(b.Sub(a).MulScalar(t))
}

func (a Color) Add(b Color) Color {
	return Color{a.R + b.R, a.G + b.G, a.B + b.B, a.A + b.A}
}

func (a Color) Sub(b Color) Color {
	return Color{a.R - b.R, a.G - b.G, a.B - b.B, a.A - b.A}
}

func (a Color) Mul(b Color) Color {
	return Color{a.R * b.R, a.G * b.G, a.B * b.B, a.A * b.A}
}

func (a Color) Div(b Color) Color {
	return Color{a.R / b.R, a.G / b.G, a.B / b.B, a.A / b.A}
}

func (a Color) AddScalar(b float64) Color {
	return Color{a.R + b, a.G + b, a.B + b, a.A + b}
}

func (a Color) SubScalar(b float64) Color {
	return Color{a.R - b, a.G - b, a.B - b, a.A - b}
}

func (a Color) MulScalar(b float64) Color {
	return Color{a.R * b, a.G * b, a.B * b, a.A * b}
}

func (a Color) DivScalar(b float64) Color {
	return Color{a.R / b, a.G / b, a.B / b, a.A / b}
}

func (a Color) Pow(b float64) Color {
	return Color{math.Pow(a.R, b), math.Pow(a.G, b), math.Pow(a.B, b), math.Pow(a.A, b)}
}

func (a Color) Min(b Color) Color {
	return Color{math.Min(a.R, b.R), math.Min(a.G, b.G), math.Min(a.B, b.B), math.Min(a.A, b.A)}
}

func (a Color) Max(b Color) Color {
	return Color{math.Max(a.R, b.R), math.Max(a.G, b.G), math.Max(a.B, b.B), math.Max(a.A, b.A)}
}
