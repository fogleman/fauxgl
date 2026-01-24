package fauxgl_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/fogleman/fauxgl"
)

//go:embed box.stl
var boxSTL []byte

//go:embed box.ascii.stl
var boxSTLA []byte

func TestLoadSTL(t *testing.T) {
	mesh, err := fauxgl.LoadSTLReader(bytes.NewReader(boxSTL), int64(len(boxSTL)))
	if err != nil {
		t.Fatalf("failed loading box.stl mesh: %s", err)
	}

	if len(mesh.Triangles) != 12 {
		t.Fatalf("failed loading triangles, expected %d", len(mesh.Triangles))
	}
}

func TestLoadSTLA(t *testing.T) {
	mesh, err := fauxgl.LoadSTLReader(bytes.NewReader(boxSTLA), int64(len(boxSTLA)))
	if err != nil {
		t.Fatalf("failed loading box.ascii.stl mesh: %s", err)
	}

	if len(mesh.Triangles) != 4 {
		t.Fatalf("failed loading triangles, expected %d", len(mesh.Triangles))
	}
}
