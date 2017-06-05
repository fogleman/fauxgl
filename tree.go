package fauxgl

import "fmt"

func NewTreeForMesh(mesh *Mesh) []Box {
	boxes := make([]Box, len(mesh.Triangles))
	for i, t := range mesh.Triangles {
		boxes[i] = t.BoundingBox()
	}
	node := NewNode(boxes)
	node.Split(0)
	return node.Leaves()
}

type Node struct {
	Box   Box
	Boxes []Box
	Axis  Axis
	Point float64
	Left  *Node
	Right *Node
}

func NewNode(boxes []Box) *Node {
	box := BoxForBoxes(boxes)
	return &Node{box, boxes, AxisNone, 0, nil, nil}
}

func (node *Node) Leaves() []Box {
	var result []Box
	if node.Left == nil && node.Right == nil {
		return []Box{node.Box}
	}
	if node.Left != nil {
		result = append(result, node.Left.Leaves()...)
	}
	if node.Right != nil {
		result = append(result, node.Right.Leaves()...)
	}
	return result
}

func (node *Node) PartitionScore(axis Axis, point float64, side bool) float64 {
	var left, right Box
	for _, box := range node.Boxes {
		l, r := box.Partition(axis, point)
		if l && r {
			if side {
				left = left.Extend(box)
			} else {
				right = right.Extend(box)
			}
		} else if l {
			left = left.Extend(box)
		} else if r {
			right = right.Extend(box)
		}
	}
	return left.Volume() + right.Volume() - left.Intersection(right).Volume()
}

func (node *Node) Partition(axis Axis, point float64, side bool) (left, right []Box) {
	for _, box := range node.Boxes {
		l, r := box.Partition(axis, point)
		if l && r {
			if side {
				left = append(left, box)
			} else {
				right = append(right, box)
			}
		} else if l {
			left = append(left, box)
		} else if r {
			right = append(right, box)
		}
	}
	return
}

func (node *Node) Split(depth int) {
	if depth >= 8 {
		return
	}
	box := node.Box
	best := box.Volume() //* 0.9
	bestAxis := AxisNone
	bestPoint := 0.0
	bestSide := false
	const N = 16
	for s := 0; s < 2; s++ {
		side := s == 1
		for i := 1; i < N; i++ {
			p := float64(i) / N
			x := box.Min.X + (box.Max.X-box.Min.X)*p
			y := box.Min.Y + (box.Max.Y-box.Min.Y)*p
			z := box.Min.Z + (box.Max.Z-box.Min.Z)*p
			sx := node.PartitionScore(AxisX, x, side)
			if sx < best {
				best = sx
				bestAxis = AxisX
				bestPoint = x
				bestSide = side
			}
			sy := node.PartitionScore(AxisY, y, side)
			if sy < best {
				best = sy
				bestAxis = AxisY
				bestPoint = y
				bestSide = side
			}
			sz := node.PartitionScore(AxisZ, z, side)
			if sz < best {
				best = sz
				bestAxis = AxisZ
				bestPoint = z
				bestSide = side
			}
		}
	}
	if bestAxis == AxisNone {
		return
	}
	l, r := node.Partition(bestAxis, bestPoint, bestSide)
	node.Axis = bestAxis
	node.Point = bestPoint
	node.Left = NewNode(l)
	node.Right = NewNode(r)
	node.Left.Split(depth + 1)
	node.Right.Split(depth + 1)
	node.Boxes = nil // only needed at leaf nodes
	left := node.Left.Box
	right := node.Right.Box
	before := box.Volume()
	after := left.Volume() + right.Volume() - left.Intersection(right).Volume()
	fmt.Println(depth, before, after, after/before)
}
