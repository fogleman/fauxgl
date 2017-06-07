package fauxgl

import (
	"math"
	"math/rand"
)

const (
	PackDetail = 5
	PackNaive  = true
)

type PackUndo struct {
	Index       int
	Rotation    Matrix
	Translation Matrix
}

type PackItem struct {
	Mesh        *Mesh
	Tree        *Node
	Rotation    Matrix
	Translation Matrix
}

func (item *PackItem) Matrix() Matrix {
	return item.Translation.Mul(item.Rotation)
}

func (item *PackItem) Copy() *PackItem {
	dup := *item
	return &dup
}

type PackModel struct {
	Items     []*PackItem
	MinVolume float64
	MaxVolume float64
}

func NewPackModel() *PackModel {
	return &PackModel{nil, 0, 0}
}

func (m *PackModel) Add(mesh *Mesh, count int) {
	detail := PackDetail
	if PackNaive {
		detail = 0
	}
	tree := NewTreeForMesh(mesh, detail)
	var offset Vector
	for i := 0; i < count; i++ {
		item := PackItem{mesh, tree, Identity(), Translate(offset)}
		m.Items = append(m.Items, &item)
		offset.X += tree.Box.Size().X + 1
		m.MinVolume = math.Max(m.MinVolume, tree.Box.Volume())
		m.MaxVolume += tree.Box.Volume()
	}
}

func (m *PackModel) Valid(i int) bool {
	item1 := m.Items[i]
	matrix1 := item1.Matrix()
	for j := 0; j < len(m.Items); j++ {
		if j == i {
			continue
		}
		item2 := m.Items[j]
		matrix2 := item2.Matrix()
		if item1.Tree.Intersects(item2.Tree, matrix1, matrix2) {
			return false
		}
	}
	return true
}

func (m *PackModel) BoundingBox() Box {
	box := EmptyBox
	for _, item := range m.Items {
		box = box.Extend(item.Tree.Box.Transform(item.Matrix()))
	}
	return box
}

func (m *PackModel) Volume() float64 {
	return m.BoundingBox().Volume()
}

func (m *PackModel) Energy() float64 {
	return m.BoundingBox().Size().MaxComponent() + math.Pow(m.Volume(), 1/3.0)
	// return (m.Volume() - m.MinVolume) / (m.MaxVolume - m.MinVolume)
	// return m.Volume() / m.MaxVolume
	// return m.Volume()
	// return math.Pow(m.Volume(), 1/3.0)
}

func (m *PackModel) DoMove() interface{} {
	i := rand.Intn(len(m.Items))
	item := m.Items[i]
	undo := PackUndo{i, item.Rotation, item.Translation}
	for {
		if !PackNaive && rand.Intn(8) == 0 {
			// rotate
			var axis Vector
			switch rand.Intn(3) {
			case 0:
				axis = Vector{1, 0, 0}
			case 1:
				axis = Vector{0, 1, 0}
			case 2:
				axis = Vector{0, 0, 1}
			}
			angle := Radians(90 * float64(rand.Intn(2)*2-1))
			item.Rotation = item.Rotation.Rotate(axis, angle)
		} else {
			// translate
			var axis Vector
			switch rand.Intn(3) {
			case 0:
				axis = Vector{1, 0, 0}
			case 1:
				axis = Vector{0, 1, 0}
			case 2:
				axis = Vector{0, 0, 1}
			}
			offset := axis
			offset = offset.MulScalar(float64(rand.Intn(2)*2 - 1))
			offset = offset.MulScalar(rand.NormFloat64() * 3)
			item.Translation = item.Translation.Translate(offset)
		}
		if m.Valid(i) {
			break
		}
		item.Rotation = undo.Rotation
		item.Translation = undo.Translation
	}
	return undo
}

func (m *PackModel) UndoMove(undo interface{}) {
	u := undo.(PackUndo)
	item := m.Items[u.Index]
	item.Rotation = u.Rotation
	item.Translation = u.Translation
}

func (m *PackModel) Copy() Annealable {
	items := make([]*PackItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = item.Copy()
	}
	return &PackModel{items, m.MinVolume, m.MaxVolume}
}
