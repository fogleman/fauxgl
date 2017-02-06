package fauxgl

var EmptyBox = Box{}

type Box struct {
	Min, Max Vector
}

func (a Box) Anchor(anchor Vector) Vector {
	return a.Min.Add(a.Size().Mul(anchor))
}

func (a Box) Center() Vector {
	return a.Anchor(Vector{0.5, 0.5, 0.5})
}

func (a Box) Size() Vector {
	return a.Max.Sub(a.Min)
}

func (a Box) Extend(b Box) Box {
	if a == EmptyBox {
		return b
	}
	return Box{a.Min.Min(b.Min), a.Max.Max(b.Max)}
}

func (a Box) Contains(b Vector) bool {
	return a.Min.X <= b.X && a.Max.X >= b.X &&
		a.Min.Y <= b.Y && a.Max.Y >= b.Y &&
		a.Min.Z <= b.Z && a.Max.Z >= b.Z
}

func (a Box) Intersects(b Box) bool {
	return !(a.Min.X > b.Max.X || a.Max.X < b.Min.X || a.Min.Y > b.Max.Y ||
		a.Max.Y < b.Min.Y || a.Min.Z > b.Max.Z || a.Max.Z < b.Min.Z)
}
