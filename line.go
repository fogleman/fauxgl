package fauxgl

type Line struct {
	V1, V2 Vertex
}

func NewLine(v1, v2 Vertex) *Line {
	return &Line{v1, v2}
}
