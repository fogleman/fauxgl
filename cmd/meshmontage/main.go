package main

import (
	"fmt"
	"log"
	"math"

	kingpin "github.com/alecthomas/kingpin/v2"
	. "github.com/fogleman/fauxgl"
)

var (
	// scale  = kingpin.Flag("scale", "Scale factor.").Short('s').Required().Float64()
	axisX  = kingpin.Flag("x", "Use X axis.").Short('x').Bool()
	axisY  = kingpin.Flag("y", "Use Y axis.").Short('y').Bool()
	axisZ  = kingpin.Flag("z", "Use Z axis.").Short('z').Bool()
	rows   = kingpin.Flag("rows", "Number of rows.").Short('r').Int()
	output = kingpin.Flag("output", "Output STL file.").Short('o').Required().String()
	files  = kingpin.Arg("files", "Models to process.").Required().ExistingFiles()
)

func main() {
	kingpin.Parse()

	var meshes []*Mesh

	for _, filename := range *files {
		fmt.Println(filename)

		mesh, err := LoadMesh(filename)
		if err != nil {
			log.Fatal(err)
		}

		meshes = append(meshes, mesh)

		// mesh.Transform(Scale(V(*scale, *scale, *scale)))

		// dir, file := filepath.Split(filename)
		// name := strings.TrimSuffix(file, filepath.Ext(file))
		// name = *prefix + name + *suffix + ".stl"

		// if *output != "" {
		// 	dir = *output
		// }

		// outputFilename := filepath.Join(dir, name)
		// fmt.Printf(" => %s\n", outputFilename)
		// if err := mesh.SaveSTL(outputFilename); err != nil {
		// 	log.Fatal(err)
		// }
	}

	maxSize := meshes[0].BoundingBox().Size()
	for _, mesh := range meshes {
		maxSize = maxSize.Max(mesh.BoundingBox().Size())
	}
	maxSize = maxSize.MulScalar(1.05)

	n := len(meshes)
	numRows := int(math.Ceil(math.Sqrt(float64(n))))
	if *rows > 0 {
		numRows = *rows
	}

	combined := NewTriangleMesh(nil)
	for i, mesh := range meshes {
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
		mesh.MoveTo(V(x, y, z), V(0.5, 0.5, 0.5))
		combined.Add(mesh)
	}

	combined.SaveSTL(*output)
}
