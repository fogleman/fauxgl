package fauxgl

type Line struct {
	V1, V2 Vertex
}

func NewLine(v1, v2 Vertex) *Line {
	return &Line{v1, v2}
}

func NewLineForPoints(p1, p2 Vector) *Line {
	v1 := Vertex{Position: p1}
	v2 := Vertex{Position: p2}
	return NewLine(v1, v2)
}

func (l *Line) BoundingBox() Box {
	min := l.V1.Position.Min(l.V2.Position)
	max := l.V1.Position.Max(l.V2.Position)
	return Box{min, max}
}

func (l *Line) Transform(matrix Matrix) {
	l.V1.Position = matrix.MulPosition(l.V1.Position)
	l.V2.Position = matrix.MulPosition(l.V2.Position)
	l.V1.Normal = matrix.MulDirection(l.V1.Normal)
	l.V2.Normal = matrix.MulDirection(l.V2.Normal)
}
