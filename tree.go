package fauxgl

func NewTreeForMesh(mesh *Mesh, depth int) *Node {
	boxes := make([]Box, len(mesh.Triangles))
	for i, t := range mesh.Triangles {
		boxes[i] = t.BoundingBox()
	}
	node := NewNode(boxes)
	node.Split(depth)
	return node
}

type Node struct {
	Box   Box
	Boxes []Box
	Left  *Node
	Right *Node
}

func NewNode(boxes []Box) *Node {
	box := BoxForBoxes(boxes)
	return &Node{box, boxes, nil, nil}
}

func (node *Node) Leaves(maxDepth int) []Box {
	var result []Box
	if maxDepth == 0 || (node.Left == nil && node.Right == nil) {
		return []Box{node.Box}
	}
	if node.Left != nil {
		result = append(result, node.Left.Leaves(maxDepth-1)...)
	}
	if node.Right != nil {
		result = append(result, node.Right.Leaves(maxDepth-1)...)
	}
	return result
}

func (node *Node) PartitionScore(axis Axis, point float64, side bool) float64 {
	var major Box
	for _, box := range node.Boxes {
		l, r := box.Partition(axis, point)
		if (l && r) || (l && side) || (r && !side) {
			major = major.Extend(box)
		}
	}
	var minor Box
	for _, box := range node.Boxes {
		if !major.ContainsBox(box) {
			minor = minor.Extend(box)
		}
	}
	return major.Volume() + minor.Volume() - major.Intersection(minor).Volume()
}

func (node *Node) Partition(axis Axis, point float64, side bool) (left, right []Box) {
	var major Box
	for _, box := range node.Boxes {
		l, r := box.Partition(axis, point)
		if (l && r) || (l && side) || (r && !side) {
			major = major.Extend(box)
		}
	}
	for _, box := range node.Boxes {
		if major.ContainsBox(box) {
			left = append(left, box)
		} else {
			right = append(right, box)
		}
	}
}

func (node *Node) Split(depth int) {
	if depth == 0 {
		return
	}
	box := node.Box
	best := box.Volume()
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
	node.Left = NewNode(l)
	node.Right = NewNode(r)
	node.Left.Split(depth - 1)
	node.Right.Split(depth - 1)
	node.Boxes = nil // only needed at leaf nodes
}
