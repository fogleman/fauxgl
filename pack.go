package fauxgl

import "math/rand"

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
	Items          []*PackItem
	UnpackedVolume float64
}

func NewPackModel() *PackModel {
	return &PackModel{nil, 0}
}

func (m *PackModel) Add(mesh *Mesh, count int) {
	tree := NewTreeForMesh(mesh, 5)
	var offset Vector
	for i := 0; i < count; i++ {
		item := PackItem{mesh, tree, Identity(), Translate(offset)}
		m.Items = append(m.Items, &item)
		offset.X += tree.Box.Size().X
		m.UnpackedVolume += tree.Box.Volume()
	}
}

func (m *PackModel) Valid() bool {
	for i := 0; i < len(m.Items); i++ {
		item1 := m.Items[i]
		matrix1 := item1.Matrix()
		for j := i + 1; j < len(m.Items); j++ {
			item2 := m.Items[j]
			matrix2 := item2.Matrix()
			if item1.Tree.Intersects(item2.Tree, matrix1, matrix2) {
				return false
			}
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
	return m.BoundingBox().Size().MaxComponent()
	// return m.Volume() / m.UnpackedVolume
}

func (m *PackModel) DoMove() interface{} {
	i := rand.Intn(len(m.Items))
	item := m.Items[i]
	undo := PackUndo{i, item.Rotation, item.Translation}
	for {
		if rand.Intn(4) == 0 {
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
			offset := axis.MulScalar(float64(rand.Intn(2)*2 - 1))
			item.Translation = item.Translation.Translate(offset)
		}
		if m.Valid() {
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
	return &PackModel{items, m.UnpackedVolume}
}
