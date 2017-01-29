package fauxgl

import "math"

var Discard = Vector{-1, -1, -1}

type Shader interface {
	Vertex(Vertex) Vertex
	Fragment(Vertex) Vector
}

type DefaultShader struct {
	Matrix  Matrix
	Light   Vector
	Camera  Vector
	Color   Vector
	Texture Texture
}

func NewDefaultShader(matrix Matrix, light, camera, color Vector) *DefaultShader {
	return &DefaultShader{matrix, light, camera, color, nil}
}

func (shader *DefaultShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *DefaultShader) Fragment(v Vertex) Vector {
	color := shader.Color
	if shader.Texture != nil {
		color = shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
	}
	diffuse := math.Max(v.Normal.Dot(shader.Light), 0)
	specular := 0.0
	if diffuse > 0 {
		camera := shader.Camera.Sub(v.Position).Normalize()
		specular = math.Max(camera.Dot(shader.Light.Negate().Reflect(v.Normal)), 0)
		specular = math.Pow(specular, 50)
	}
	light := Clamp(diffuse+specular, 0.1, 1)
	return color.MulScalar(light)
}
