package soft

var Discard = Vector{-1, -1, -1}

type Shader interface {
	Vertex(Vertex) Vertex
	Fragment(Vertex) Vector
}

type DefaultShader struct {
	Matrix Matrix
	Light  Vector
	Color  Vector
}

func NewDefaultShader(matrix Matrix, light, color Vector) Shader {
	return &DefaultShader{matrix, light, color}
}

func (shader *DefaultShader) Vertex(v Vertex) Vertex {
	v.Position = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *DefaultShader) Fragment(v Vertex) Vector {
	light := Clamp(v.Normal.Dot(shader.Light), 0, 1)
	return shader.Color.MulScalar(light)
}
