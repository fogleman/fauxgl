package fauxgl

import (
	"fmt"
	"testing"
)

func TestHidden(t *testing.T) {
	// p0 := V(-21, -72, 63)
	// p1 := V(-78, 99, 40)
	// p2 := V(-19, -78, -83)
	// q0 := V(96, 77, -51)
	// q1 := V(-95, -1, -16)
	// q2 := V(9, 5, -21)
	p0 := V(0, 0, 0)
	p1 := V(1, 0, 0)
	p2 := V(0, 1, 0)
	q0 := V(0.5, 0, -0.5)
	q1 := V(0.5, 0, 0.5)
	q2 := V(0.5, 1, 0.5)

	x0, x1, ok := triangleTriangleIntersection(p0, p1, p2, q0, q1, q2)
	fmt.Println(x0, x1, ok)

	t0, t1, ok := clipSegment(p0, p1, p2, q0, q1, q2)
	fmt.Println(t0, t1, ok)

	t0, t1, ok = clipSegment(q0, q1, q2, p0, p1, p2)
	fmt.Println(t0, t1, ok)

	// func triangleTriangleIntersection(p00, p01, p02, p10, p11, p12 Vector) (Vector, Vector, bool) {
}
