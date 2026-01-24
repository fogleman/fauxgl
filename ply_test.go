package fauxgl_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/fogleman/fauxgl"
)

//go:embed pyramid.ply
var pyramidPLY []byte

func TestLoadPLY(t *testing.T) {
	mesh, err := fauxgl.LoadPLYReader(bytes.NewReader(pyramidPLY))
	if err != nil {
		t.Fatalf("failed loading box.ply mesh: %s", err)
	}

	if len(mesh.Triangles) != 6 {
		t.Fatalf("failed loading triangles, expected %d", len(mesh.Triangles))
	}
}
