package main

import (
	"fmt"
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
	var done func()

	done = timed("loading mesh")
	mesh, err := LoadMesh(os.Args[1])
	if err != nil {
		panic(err)
	}
	done()

	done = timed("writing stl")
	mesh.SaveSTL(os.Args[2])
	done()
}
