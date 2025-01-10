package main

import (
	"log"
	"math"

	kingpin "github.com/alecthomas/kingpin/v2"
	. "github.com/fogleman/fauxgl"
)

var (
	// scale  = kingpin.Flag("scale", "Scale factor.").Short('s').Required().Float64()
	count  = kingpin.Flag("count", "Number of repeats.").Short('n').Required().Int()
	axisX  = kingpin.Flag("x", "Use X axis.").Short('x').Bool()
	axisY  = kingpin.Flag("y", "Use Y axis.").Short('y').Bool()
	axisZ  = kingpin.Flag("z", "Use Z axis.").Short('z').Bool()
	rows   = kingpin.Flag("rows", "Number of rows.").Short('r').Int()
	output = kingpin.Flag("output", "Output STL file.").Short('o').Required().String()
	file   = kingpin.Arg("file", "Model to repeat.").Required().ExistingFile()
)

func main() {
	kingpin.Parse()

	mesh, err := LoadMesh(*file)
	if err != nil {
		log.Fatal(err)
	}

	maxSize := mesh.BoundingBox().Size()
	maxSize = maxSize.MulScalar(2)

	numRows := int(math.Ceil(math.Sqrt(float64(*count))))
	if *rows > 0 {
		numRows = *rows
	}

	combined := NewTriangleMesh(nil)
	for i := 0; i < *count; i++ {
		row := i / numRows
		col := i % numRows
		var x, y, z float64
		if *axisX {
			y = float64(row) * maxSize.Y
			z = float64(col) * maxSize.Z
		} else if *axisY {
			x = float64(row) * maxSize.X
			z = float64(col) * maxSize.Z
		} else {
			x = float64(row) * maxSize.X
			y = float64(col) * maxSize.Y
		}
		m := mesh.Copy()
		m.MoveTo(V(x, y, z), V(0.5, 0.5, 0.5))
		combined.Add(m)
	}

	combined.MoveTo(V(0, 0, 0), V(0.5, 0.5, 0))

	combined.SaveSTL(*output)
}
