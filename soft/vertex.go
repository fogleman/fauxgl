package soft

type Vertex struct {
	Position Vector
	Normal   Vector
	Texture  Vector
	Color    Vector
	Output   VectorW
}

func (a Vertex) Outside() bool {
	return a.Output.Outside()
}

var zeroVector = Vector{}
var zeroVectorW = VectorW{}

func InterpolateVertexes(v1, v2, v3 Vertex, b Vector) Vertex {
	v := Vertex{}
	v.Position = InterpolateVectors(v1.Position, v2.Position, v3.Position, b)
	v.Normal = InterpolateVectors(v1.Normal, v2.Normal, v3.Normal, b).Normalize()
	v.Texture = InterpolateVectors(v1.Texture, v2.Texture, v3.Texture, b)
	v.Color = InterpolateVectors(v1.Color, v2.Color, v3.Color, b)
	v.Output = InterpolateVectorWs(v1.Output, v2.Output, v3.Output, b)
	return v
}

func InterpolateVectors(v1, v2, v3, b Vector) Vector {
	if v1 == zeroVector && v2 == zeroVector && v3 == zeroVector {
		return zeroVector
	}
	n := Vector{}
	n = n.Add(v1.MulScalar(b.X))
	n = n.Add(v2.MulScalar(b.Y))
	n = n.Add(v3.MulScalar(b.Z))
	return n
}

func InterpolateVectorWs(v1, v2, v3 VectorW, b Vector) VectorW {
	if v1 == zeroVectorW && v2 == zeroVectorW && v3 == zeroVectorW {
		return zeroVectorW
	}
	n := VectorW{}
	n = n.Add(v1.MulScalar(b.X))
	n = n.Add(v2.MulScalar(b.Y))
	n = n.Add(v3.MulScalar(b.Z))
	return n
}

func Barycentric(p1, p2, p3, p Vector) Vector {
	v0 := p2.Sub(p1)
	v1 := p3.Sub(p1)
	v2 := p.Sub(p1)
	d00 := v0.Dot(v0)
	d01 := v0.Dot(v1)
	d11 := v1.Dot(v1)
	d20 := v2.Dot(v0)
	d21 := v2.Dot(v1)
	d := d00*d11 - d01*d01
	v := (d11*d20 - d01*d21) / d
	w := (d00*d21 - d01*d20) / d
	u := 1 - v - w
	return Vector{u, v, w}
}
