package fauxgl

type Edge struct {
	A, B Vector
}

func EdgeKey(a, b Vector) Edge {
	if a.Less(b) {
		return Edge{a, b}
	} else {
		return Edge{b, a}
	}
}

func (e Edge) Key() Edge {
	return EdgeKey(e.A, e.B)
}
