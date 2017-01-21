package soft

import (
	"image"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"strconv"
)

func Radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func Degrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	im, _, err := image.Decode(file)
	return im, err
}

func SavePNG(path string, im image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}

func ParseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}

func ClampInt(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
