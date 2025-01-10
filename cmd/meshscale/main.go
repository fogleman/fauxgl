package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	kingpin "github.com/alecthomas/kingpin/v2"
	. "github.com/fogleman/fauxgl"
)

var (
	scale  = kingpin.Flag("scale", "Scale factor.").Short('s').Required().Float64()
	prefix = kingpin.Flag("prefix", "Prefix to apply to each output filename.").Default("").String()
	suffix = kingpin.Flag("suffix", "Suffix to apply to each output filename.").Default("").String()
	output = kingpin.Flag("output", "Output directory.").Short('o').Default("").String()
	files  = kingpin.Arg("files", "Models to process.").Required().ExistingFiles()
)

func main() {
	kingpin.Parse()

	if *output != "" {
		if err := os.MkdirAll(*output, 0755); err != nil {
			log.Fatal(err)
		}
	}

	for _, filename := range *files {
		fmt.Println(filename)

		mesh, err := LoadMesh(filename)
		if err != nil {
			log.Fatal(err)
		}

		mesh.Transform(Scale(V(*scale, *scale, *scale)))

		dir, file := filepath.Split(filename)
		name := strings.TrimSuffix(file, filepath.Ext(file))
		name = *prefix + name + *suffix + ".stl"

		if *output != "" {
			dir = *output
		}

		outputFilename := filepath.Join(dir, name)
		fmt.Printf(" => %s\n", outputFilename)
		if err := mesh.SaveSTL(outputFilename); err != nil {
			log.Fatal(err)
		}
	}
}
