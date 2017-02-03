package fauxgl

type Voxel struct {
	X, Y, Z int
	Color   Color
}

func (v Voxel) key(dx, dy, dz int) Voxel {
	return Voxel{v.X + dx, v.Y + dy, v.Z + dz, Color{}}
}

type voxelAxis int

const (
	_ voxelAxis = iota
	voxelX
	voxelY
	voxelZ
)

type voxelNormal struct {
	Axis voxelAxis
	Sign int
}

var (
	voxelPosX = voxelNormal{voxelX, 1}
	voxelNegX = voxelNormal{voxelX, -1}
	voxelPosY = voxelNormal{voxelY, 1}
	voxelNegY = voxelNormal{voxelY, -1}
	voxelPosZ = voxelNormal{voxelZ, 1}
	voxelNegZ = voxelNormal{voxelZ, -1}
)

type voxelPlane struct {
	Normal   voxelNormal
	Position int
	Color    Color
}

type voxelFace struct {
	I0, J0 int
	I1, J1 int
}

func NewVoxelMesh(voxels []Voxel) *Mesh {
	lookup := make(map[Voxel]bool)
	for _, v := range voxels {
		lookup[v.key(0, 0, 0)] = true
	}

	planeFaces := make(map[voxelPlane][]voxelFace)
	for _, v := range voxels {
		if !lookup[v.key(0, 0, 1)] {
			plane := voxelPlane{voxelPosZ, v.Z, v.Color}
			face := voxelFace{v.X, v.Y, v.X, v.Y}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[v.key(0, 0, -1)] {
			plane := voxelPlane{voxelNegZ, v.Z, v.Color}
			face := voxelFace{v.X, v.Y, v.X, v.Y}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[v.key(0, -1, 0)] {
			plane := voxelPlane{voxelNegY, v.Y, v.Color}
			face := voxelFace{v.X, v.Z, v.X, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[v.key(1, 0, 0)] {
			plane := voxelPlane{voxelPosX, v.X, v.Color}
			face := voxelFace{v.Y, v.Z, v.Y, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[v.key(0, 1, 0)] {
			plane := voxelPlane{voxelPosY, v.Y, v.Color}
			face := voxelFace{v.X, v.Z, v.X, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[v.key(-1, 0, 0)] {
			plane := voxelPlane{voxelNegX, v.X, v.Color}
			face := voxelFace{v.Y, v.Z, v.Y, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
	}

	var triangles []*Triangle
	for plane, faces := range planeFaces {
		faces = combineVoxelFaces(faces)
		k := float64(plane.Position) + float64(plane.Normal.Sign)*0.5
		for _, face := range faces {
			i0 := float64(face.I0) - 0.5
			j0 := float64(face.J0) - 0.5
			i1 := float64(face.I1) + 0.5
			j1 := float64(face.J1) + 0.5
			var p1, p2, p3, p4 Vector
			switch plane.Normal.Axis {
			case voxelX:
				p1 = Vector{k, i0, j0}
				p2 = Vector{k, i1, j0}
				p3 = Vector{k, i1, j1}
				p4 = Vector{k, i0, j1}
			case voxelY:
				p1 = Vector{i0, k, j1}
				p2 = Vector{i1, k, j1}
				p3 = Vector{i1, k, j0}
				p4 = Vector{i0, k, j0}
			case voxelZ:
				p1 = Vector{i0, j0, k}
				p2 = Vector{i1, j0, k}
				p3 = Vector{i1, j1, k}
				p4 = Vector{i0, j1, k}
			}
			if plane.Normal.Sign < 0 {
				p1, p2, p3, p4 = p4, p3, p2, p1
			}
			t1 := NewTriangleForPoints(p1, p2, p3)
			t2 := NewTriangleForPoints(p1, p3, p4)
			t1.V1.Color = plane.Color
			t1.V2.Color = plane.Color
			t1.V3.Color = plane.Color
			t2.V1.Color = plane.Color
			t2.V2.Color = plane.Color
			t2.V3.Color = plane.Color
			triangles = append(triangles, t1)
			triangles = append(triangles, t2)
		}
	}

	return NewMesh(triangles)
}

func combineVoxelFaces(faces []voxelFace) []voxelFace {
	var result []voxelFace
	for len(faces) > 0 {
		face := largestRectangle(faces)
		result = append(result, face)
		var newFaces []voxelFace
		for _, f := range faces {
			if f.I0 >= face.I0 && f.I1 <= face.I1 && f.J0 >= face.J0 && f.J1 <= face.J1 {
				continue
			}
			newFaces = append(newFaces, f)
		}
		faces = newFaces
	}
	return result
}

func largestRectangle(faces []voxelFace) voxelFace {
	// determine bounding box
	i0 := faces[0].I0
	j0 := faces[0].J0
	i1 := faces[0].I1
	j1 := faces[0].J1
	for _, f := range faces {
		if f.I0 < i0 {
			i0 = f.I0
		}
		if f.J0 < j0 {
			j0 = f.J0
		}
		if f.I1 > i1 {
			i1 = f.I1
		}
		if f.J1 > j1 {
			j1 = f.J1
		}
	}
	// create arrays
	nj := j1 - j0 + 1
	ni := i1 - i0 + 1
	a := make([][]int, nj)
	w := make([][]int, nj)
	h := make([][]int, nj)
	for j := range a {
		a[j] = make([]int, ni)
		w[j] = make([]int, ni)
		h[j] = make([]int, ni)
	}
	// populate array
	for _, f := range faces {
		for j := f.J0; j <= f.J1; j++ {
			for i := f.I0; i <= f.I1; i++ {
				a[j-j0][i-i0] = 1
			}
		}
	}
	// find largest area rectangle
	var maxArea int
	var maxFace voxelFace
	for j := 0; j < nj; j++ {
		for i := 0; i < ni; i++ {
			if a[j][i] == 0 {
				continue
			}
			if j == 0 {
				h[j][i] = 1
			} else {
				h[j][i] = h[j-1][i] + 1
			}
			if i == 0 {
				w[j][i] = 1
			} else {
				w[j][i] = w[j][i-1] + 1
			}
			minw := w[j][i]
			for dh := 0; dh < h[j][i]; dh++ {
				if w[j-dh][i] < minw {
					minw = w[j-dh][i]
				}
				area := (dh + 1) * minw
				if area > maxArea {
					maxArea = area
					maxFace = voxelFace{i0 + i - minw + 1, j0 + j - dh, i0 + i, j0 + j}
				}
			}
		}
	}
	return maxFace
}
