package soft

func Rasterize(p1, p2, p3 Vector, buffer []Fragment) []Fragment {
	min := p1.Min(p2.Min(p3)).Floor()
	max := p1.Max(p2.Max(p3)).Ceil()
	x1 := int(min.X)
	x2 := int(max.X)
	y1 := int(min.Y)
	y2 := int(max.Y)
	fragments := buffer[:0]
	v0 := p2.Sub(p1)
	v1 := p3.Sub(p1)
	d00 := v0.X*v0.X + v0.Y*v0.Y
	d01 := v0.X*v1.X + v0.Y*v1.Y
	d11 := v1.X*v1.X + v1.Y*v1.Y
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			p := Vector{float64(x) + 0.5, float64(y) + 0.5, 0}
			v2 := p.Sub(p1)
			d20 := v2.X*v0.X + v2.Y*v0.Y
			d21 := v2.X*v1.X + v2.Y*v1.Y
			d := d00*d11 - d01*d01
			v := (d11*d20 - d01*d21) / d
			if v < 0 {
				continue
			}
			w := (d00*d21 - d01*d20) / d
			if w < 0 {
				continue
			}
			u := 1 - v - w
			if u < 0 {
				continue
			}
			b := Vector{u, v, w}
			z := b.X*p1.Z + b.Y*p2.Z + b.Z*p3.Z
			f := Fragment{Vector{float64(x), float64(y), z}, b}
			fragments = append(fragments, f)
		}
	}
	return fragments
}
