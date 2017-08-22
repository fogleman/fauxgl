package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	. "github.com/fogleman/fauxgl"
)

func timed(name string) func() {
	if len(name) > 0 {
		fmt.Printf("%s... ", name)
	}
	start := time.Now()
	return func() {
		fmt.Println(time.Since(start))
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var done func()

	done = timed("loading mesh")
	mesh, err := LoadMesh(os.Args[1])
	if err != nil {
		panic(err)
	}
	done()

	fmt.Println("original volume is", mesh.BoundingBox().Volume())

	// a := float64(rand.Intn(360) - 180)
	// fmt.Println(a)
	// mesh.Transform(Rotate(Vector{0, 0, 1}, Radians(a)))
	// // mesh.Transform(RotateTo(RandomUnitVector(), RandomUnitVector()))
	mesh.SaveSTL("out1.stl")

	done = timed("aligning mesh")
	mesh.AxisAlignZ()
	// mesh.AxisAlign(Vector{0, 1, 0})
	// mesh.AxisAlign(Vector{1, 0, 0})
	// mesh.AxisAlign(Vector{0, 0, 1})
	// mesh.AxisAlign(Vector{0, 1, 0})
	// mesh.AxisAlign(Vector{1, 0, 0})
	done()

	fmt.Println("aligned volume is", mesh.BoundingBox().Volume())

	// done = timed("writing output")
	mesh.SaveSTL("out2.stl")
	// done()
}
