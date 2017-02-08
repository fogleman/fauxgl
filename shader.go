package fauxgl

import "math"

type Shader interface {
	Vertex(Vertex) Vertex
	Fragment(Vertex) Color
}

type DefaultShader struct {
	Matrix  Matrix
	Light   Vector
	Camera  Vector
	Color   Color
	Texture Texture
}

func NewDefaultShader(matrix Matrix, light, camera Vector, color Color) *DefaultShader {
	return &DefaultShader{matrix, light, camera, color, nil}
}

func (shader *DefaultShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *DefaultShader) Fragment(v Vertex) Color {
	color := shader.Color
	if color == Discard {
		color = v.Color
	}
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
	return color.MulScalar(light).Alpha(color.A)
}

// SolidColorShader renders with a single, solid color.
type SolidColorShader struct {
	Matrix Matrix
	Color  Color
}

func NewSolidColorShader(matrix Matrix, color Color) *SolidColorShader {
	return &SolidColorShader{matrix, color}
}

func (shader *SolidColorShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *SolidColorShader) Fragment(v Vertex) Color {
	return shader.Color
}

// TextureShader renders with a texture and no lighting.
type TextureShader struct {
	Matrix  Matrix
	Texture Texture
}

func NewTextureShader(matrix Matrix, texture Texture) *TextureShader {
	return &TextureShader{matrix, texture}
}

func (shader *TextureShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *TextureShader) Fragment(v Vertex) Color {
	return shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
}

// PhongShader implements Phong shading with an optional texture.
type PhongShader struct {
	Matrix         Matrix
	LightDirection Vector
	CameraPosition Vector
	ObjectColor    Color
	AmbientColor   Color
	DiffuseColor   Color
	SpecularColor  Color
	Texture        Texture
	SpecularPower  float64
}

func NewPhongShader(matrix Matrix, lightDirection, cameraPosition Vector) *PhongShader {
	ambient := Color{0.2, 0.2, 0.2, 1}
	diffuse := Color{0.8, 0.8, 0.8, 1}
	specular := Color{1, 1, 1, 1}
	return &PhongShader{
		matrix, lightDirection, cameraPosition,
		Discard, ambient, diffuse, specular, nil, 32}
}

func (shader *PhongShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *PhongShader) Fragment(v Vertex) Color {
	light := shader.AmbientColor
	color := v.Color
	if shader.ObjectColor != Discard {
		color = shader.ObjectColor
	}
	if shader.Texture != nil {
		color = shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
	}
	diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	light = light.Add(shader.DiffuseColor.MulScalar(diffuse))
	if diffuse > 0 {
		camera := shader.CameraPosition.Sub(v.Position).Normalize()
		reflected := shader.LightDirection.Negate().Reflect(v.Normal)
		specular := math.Max(camera.Dot(reflected), 0)
		if specular > 0 {
			specular = math.Pow(specular, shader.SpecularPower)
			light = light.Add(shader.SpecularColor.MulScalar(specular))
		}
	}
	return color.Mul(light).Min(White).Alpha(color.A)
}
