package soft

type Vertex struct {
	Position Vector
	Normal   Vector
	Texture  Vector
	Color    Vector
}

var zeroVector = Vector{}

func InterpolateVertexes(v1, v2, v3 Vertex, b Vector) Vertex {
	v := Vertex{}
	v.Position = InterpolateVectors(v1.Position, v2.Position, v3.Position, b)
	v.Normal = InterpolateVectors(v1.Normal, v2.Normal, v3.Normal, b).Normalize()
	v.Texture = InterpolateVectors(v1.Texture, v2.Texture, v3.Texture, b)
	v.Color = InterpolateVectors(v1.Color, v2.Color, v3.Color, b)
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
