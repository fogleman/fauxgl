package fauxgl_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/fogleman/fauxgl"
)

//go:embed box.obj
var boxOBJ []byte

func TestLoadOBJ(t *testing.T) {
	mesh, err := fauxgl.LoadOBJReader(bytes.NewReader(boxOBJ))
	if err != nil {
		t.Fatalf("failed loading box.obj mesh: %s", err)
	}

	if len(mesh.Triangles) != 12 {
		t.Fatalf("failed loading triangles, expected %d", len(mesh.Triangles))
	}
}


//go:embed pyramid.obj
var pyramidOBJ []byte

func TestLoadOBJBad(t *testing.T) {
	_, err := fauxgl.LoadOBJReader(bytes.NewReader(pyramidOBJ))
	if err == nil {
		t.Fatalf("expected failure loading pyramid.obj mesh: %s", err)
	}
}

