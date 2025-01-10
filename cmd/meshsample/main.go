package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	kingpin "github.com/alecthomas/kingpin/v2"
	. "github.com/fogleman/fauxgl"
)

var (
	count = kingpin.Flag("count", "Approximate number of points.").Short('n').Required().Int()
	file  = kingpin.Arg("file", "Model to process.").Required().ExistingFile()
)

func main() {
	kingpin.Parse()

	mesh, err := LoadMesh(*file)
	if err != nil {
		log.Fatal(err)
	}

	var totalArea float64
	for _, t := range mesh.Triangles {
		totalArea += t.Area()
	}

	countPerUnitArea := float64(*count) / totalArea
	for _, t := range mesh.Triangles {
		integer, frac := math.Modf(t.Area() * countPerUnitArea)
		n := int(integer)
		if rand.Float64() < frac {
			n++
		}
		for i := 0; i < n; i++ {
			r1 := rand.Float64()
			r2 := rand.Float64()
			s1 := math.Sqrt(r1)
			v1 := t.V1.Position.MulScalar(1 - s1)
			v2 := t.V2.Position.MulScalar(s1 * (1 - r2))
			v3 := t.V3.Position.MulScalar(r2 * s1)
			v := v1.Add(v2).Add(v3)
			fmt.Printf("%g,%g,%g\n", v.X, v.Y, v.Z)
		}
	}
}
