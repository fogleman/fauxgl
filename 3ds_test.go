package fauxgl_test

import (
	"bytes"
	_ "embed"
	"io"
	"testing"

	"github.com/fogleman/fauxgl"
)

//go:embed box.3ds
var box3DS []byte

func TestLoad3DS(t *testing.T) {
	mesh, err := fauxgl.Load3DSReader(bytes.NewReader(box3DS))
	if err != nil {
		t.Fatalf("failed loading box.3ds mesh: %s", err)
	}

	if len(mesh.Triangles) != 12 {
		t.Fatalf("failed loading triangles, expected %d", len(mesh.Triangles))
	}
}

type noseek struct {
	rs io.ReadSeeker
}

func (ns noseek) Read(p []byte) (n int, err error) {
	return ns.rs.Read(p)
}

func TestLoad3DSNoSeek(t *testing.T) {
	mesh, err := fauxgl.Load3DSReader(noseek{bytes.NewReader(box3DS)})
	if err != nil {
		t.Fatalf("failed loading box.3ds mesh: %s", err)
	}

	if len(mesh.Triangles) != 12 {
		t.Fatalf("failed loading triangles, expected %d", len(mesh.Triangles))
	}
}
