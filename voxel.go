package fauxgl

import (
	"math"
	"sort"
)

type Voxel struct {
	X, Y, Z int
	Color   Color
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

type voxelPolygon struct {
	Points []Vector
	Color  Color
}

func NewVoxelMesh(voxels []Voxel) *Mesh {
	type key struct {
		X, Y, Z int
	}

	// create lookup table
	lookup := make(map[key]bool)
	for _, v := range voxels {
		lookup[key{v.X, v.Y, v.Z}] = true
	}

	// find exposed faces
	planeFaces := make(map[voxelPlane][]voxelFace)
	for _, v := range voxels {
		if !lookup[key{v.X + 1, v.Y, v.Z}] {
			plane := voxelPlane{voxelPosX, v.X, v.Color}
			face := voxelFace{v.Y, v.Z, v.Y, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X - 1, v.Y, v.Z}] {
			plane := voxelPlane{voxelNegX, v.X, v.Color}
			face := voxelFace{v.Y, v.Z, v.Y, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y + 1, v.Z}] {
			plane := voxelPlane{voxelPosY, v.Y, v.Color}
			face := voxelFace{v.X, v.Z, v.X, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y - 1, v.Z}] {
			plane := voxelPlane{voxelNegY, v.Y, v.Color}
			face := voxelFace{v.X, v.Z, v.X, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y, v.Z + 1}] {
			plane := voxelPlane{voxelPosZ, v.Z, v.Color}
			face := voxelFace{v.X, v.Y, v.X, v.Y}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y, v.Z - 1}] {
			plane := voxelPlane{voxelNegZ, v.Z, v.Color}
			face := voxelFace{v.X, v.Y, v.X, v.Y}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
	}

	// find large rectangular regions
	for plane, faces := range planeFaces {
		planeFaces[plane] = combineVoxelFaces(faces)
	}

	// generate wireframe outlines
	var lines []*Line
	for plane, faces := range planeFaces {
		lines = append(lines, outlineVoxelFaces(plane, faces)...)
	}

	// create polygonal faces
	polygons := voxelPolygons(planeFaces)

	// segment faces at t-junctions (optional)
	// polygons = segmentVoxelPolygons(polygons)

	// triangulate polygons
	var triangles []*Triangle
	for _, p := range polygons {
		triangles = append(triangles, triangulateVoxelPolygon(p)...)
	}

	return NewMesh(triangles, lines)
}

func triangulateVoxelPolygon(polygon voxelPolygon) []*Triangle {
	p := polygon.Points
	if len(p) < 3 {
		return nil
	} else if len(p) == 3 {
		t := NewTriangleForPoints(p[0], p[1], p[2])
		t.V1.Color = polygon.Color
		t.V2.Color = polygon.Color
		t.V3.Color = polygon.Color
		return []*Triangle{t}
	} else {
		var bestI, bestJ int
		var bestAspect float64
		for i := 0; i < len(p); i++ {
			for j := i + 1; j < len(p); j++ {
				points1, points2 := voxelPolygonSplit(p, i, j)
				aspect1 := voxelPolygonAspect(points1)
				aspect2 := voxelPolygonAspect(points2)
				if len(points1) < 3 || len(points2) < 3 {
					continue
				}
				if aspect1 == math.MaxFloat64 || aspect2 == math.MaxFloat64 {
					continue
				}
				aspect := math.Min(aspect1, aspect2)
				if bestAspect == 0 || aspect < bestAspect {
					bestI = i
					bestJ = j
					bestAspect = aspect
				}
			}
		}
		points1, points2 := voxelPolygonSplit(p, bestI, bestJ)
		p1 := voxelPolygon{points1, polygon.Color}
		p2 := voxelPolygon{points2, polygon.Color}
		t1 := triangulateVoxelPolygon(p1)
		t2 := triangulateVoxelPolygon(p2)
		return append(t1, t2...)
	}
}

func voxelPolygonSplit(points []Vector, i, j int) ([]Vector, []Vector) {
	points1 := points[i : j+1]
	var points2 []Vector
	k := j
	points2 = append(points2, points[k])
	for k != i {
		k = (k + 1) % len(points)
		points2 = append(points2, points[k])
	}
	return points1, points2
}

func voxelPolygonAspect(points []Vector) float64 {
	min := points[0]
	max := points[0]
	for _, point := range points {
		min = min.Min(point)
		max = max.Max(point)
	}
	d := max.Sub(min)
	var di, dj float64
	if d.X == 0 {
		di, dj = d.Y, d.Z
	} else if d.Y == 0 {
		di, dj = d.X, d.Z
	} else if d.Z == 0 {
		di, dj = d.X, d.Y
	}
	if di < dj {
		di, dj = dj, di
	}
	if dj == 0 {
		return math.MaxFloat64
	}
	return di / dj
}

func combineVoxelFaces(faces []voxelFace) []voxelFace {
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
	count := 0
	for _, f := range faces {
		for j := f.J0; j <= f.J1; j++ {
			for i := f.I0; i <= f.I1; i++ {
				a[j-j0][i-i0] = 1
				count++
			}
		}
	}
	// find rectangles
	var result []voxelFace
	for count > 0 {
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
						maxFace = voxelFace{
							i0 + i - minw + 1, j0 + j - dh, i0 + i, j0 + j}
					}
				}
			}
		}
		result = append(result, maxFace)
		for j := maxFace.J0; j <= maxFace.J1; j++ {
			for i := maxFace.I0; i <= maxFace.I1; i++ {
				a[j-j0][i-i0] = 0
				count--
			}
		}
		for j := 0; j < nj; j++ {
			for i := 0; i < ni; i++ {
				w[j][i] = 0
				h[j][i] = 0
			}
		}
	}
	return result
}

func distinctAndSorted(a []float64) []float64 {
	sort.Float64s(a)
	b := a[:0]
	for i, x := range a {
		if i == 0 || x != a[i-1] {
			b = append(b, x)
		}
	}
	return b
}

func segmentVoxelPolygons(polygons []voxelPolygon) []voxelPolygon {
	lookup := make(map[Vector][]float64)
	any := math.MaxFloat64
	for _, p := range polygons {
		for _, v := range p.Points {
			kx := Vector{any, v.Y, v.Z}
			ky := Vector{v.X, any, v.Z}
			kz := Vector{v.X, v.Y, any}
			lookup[kx] = append(lookup[kx], v.X)
			lookup[ky] = append(lookup[ky], v.Y)
			lookup[kz] = append(lookup[kz], v.Z)
		}
	}
	for k, v := range lookup {
		lookup[k] = distinctAndSorted(v)
	}
	result := make([]voxelPolygon, len(polygons))
	for idx, p := range polygons {
		var newPoints []Vector
		for i, p1 := range p.Points {
			newPoints = append(newPoints, p1)
			p2 := p.Points[(i+1)%len(p.Points)]
			if p1.X == p2.X && p1.Y == p2.Y {
				zs := lookup[Vector{p1.X, p1.Y, any}]
				if p2.Z > p1.Z {
					for j := 0; j < len(zs); j++ {
						if zs[j] > p1.Z && zs[j] < p2.Z {
							newPoints = append(newPoints, Vector{p1.X, p1.Y, zs[j]})
						}
					}
				} else {
					for j := len(zs) - 1; j >= 0; j-- {
						if zs[j] > p2.Z && zs[j] < p1.Z {
							newPoints = append(newPoints, Vector{p1.X, p1.Y, zs[j]})
						}
					}
				}
			} else if p1.X == p2.X && p1.Z == p2.Z {
				ys := lookup[Vector{p1.X, any, p1.Z}]
				if p2.Y > p1.Y {
					for j := 0; j < len(ys); j++ {
						if ys[j] > p1.Y && ys[j] < p2.Y {
							newPoints = append(newPoints, Vector{p1.X, ys[j], p1.Z})
						}
					}
				} else {
					for j := len(ys) - 1; j >= 0; j-- {
						if ys[j] > p2.Y && ys[j] < p1.Y {
							newPoints = append(newPoints, Vector{p1.X, ys[j], p1.Z})
						}
					}
				}
			} else if p1.Y == p2.Y && p1.Z == p2.Z {
				xs := lookup[Vector{any, p1.Y, p1.Z}]
				if p2.X > p1.X {
					for j := 0; j < len(xs); j++ {
						if xs[j] > p1.X && xs[j] < p2.X {
							newPoints = append(newPoints, Vector{xs[j], p1.Y, p1.Z})
						}
					}
				} else {
					for j := len(xs) - 1; j >= 0; j-- {
						if xs[j] > p2.X && xs[j] < p1.X {
							newPoints = append(newPoints, Vector{xs[j], p1.Y, p1.Z})
						}
					}
				}
			}
		}
		p.Points = newPoints
		result[idx] = p
	}
	return result
}

func voxelPolygons(planeFaces map[voxelPlane][]voxelFace) []voxelPolygon {
	var result []voxelPolygon
	for plane, faces := range planeFaces {
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
			points := []Vector{p1, p2, p3, p4}
			polygon := voxelPolygon{points, plane.Color}
			result = append(result, polygon)
		}
	}
	return result
}

func outlineVoxelFaces(plane voxelPlane, faces []voxelFace) []*Line {
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
	// padding
	i0--
	j0--
	i1++
	j1++
	// create array
	nj := j1 - j0 + 1
	ni := i1 - i0 + 1
	a := make([][]bool, nj)
	for j := range a {
		a[j] = make([]bool, ni)
	}
	// populate array
	for _, f := range faces {
		for j := f.J0; j <= f.J1; j++ {
			for i := f.I0; i <= f.I1; i++ {
				a[j-j0][i-i0] = true
			}
		}
	}
	var lines []*Line
	for sign := -1; sign <= 1; sign += 2 {
		// find "horizontal" lines
		for j := 1; j < nj-1; j++ {
			start := -1
			for i := 0; i < ni; i++ {
				if a[j][i] && !a[j+sign][i] {
					if start < 0 {
						start = i
					}
				} else if start >= 0 {
					end := i - 1
					ai := float64(i0+start) - 0.5
					bi := float64(i0+end) + 0.5
					jj := float64(j0+j) + 0.5*float64(sign)
					line := createVoxelOutline(plane, ai, jj, bi, jj)
					lines = append(lines, line)
					start = -1
				}
			}

		}
		// find "vertical" lines
		for i := 1; i < ni-1; i++ {
			start := -1
			for j := 0; j < nj; j++ {
				if a[j][i] && !a[j][i+sign] {
					if start < 0 {
						start = j
					}
				} else if start >= 0 {
					end := j - 1
					aj := float64(j0+start) - 0.5
					bj := float64(j0+end) + 0.5
					ii := float64(i0+i) + 0.5*float64(sign)
					line := createVoxelOutline(plane, ii, aj, ii, bj)
					lines = append(lines, line)
					start = -1
				}
			}
		}
	}
	return lines
}

func createVoxelOutline(plane voxelPlane, i0, j0, i1, j1 float64) *Line {
	k := float64(plane.Position) + float64(plane.Normal.Sign)*0.5
	var p1, p2 Vector
	switch plane.Normal.Axis {
	case voxelX:
		p1 = Vector{k, i0, j0}
		p2 = Vector{k, i1, j1}
	case voxelY:
		p1 = Vector{i0, k, j0}
		p2 = Vector{i1, k, j1}
	case voxelZ:
		p1 = Vector{i0, j0, k}
		p2 = Vector{i1, j1, k}
	}
	return NewLineForPoints(p1, p2)
}
