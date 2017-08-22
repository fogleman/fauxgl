package fauxgl

import (
	"math"
	"sort"
)

func cross2D(p, a, b Vector) float64 {
	return (a.X-p.X)*(b.Y-p.Y) - (a.Y-p.Y)*(b.X-p.X)
}

func intersectLines(p0, d0, p1, d1 Vector) Vector {
	d := d0.X*d1.Y - d0.Y*d1.X
	dx := p1.X - p0.X
	dy := p1.Y - p0.Y
	t := (dx*d1.Y - dy*d1.X) / d
	return Vector{p0.X + t*d0.X, p0.Y + t*d0.Y, 0}
}

func orientedBoxArea(lp, ld, rp, rd, tp, td, bp, bd Vector) float64 {
	upperLeft := intersectLines(lp, ld, tp, td)
	upperRight := intersectLines(rp, rd, tp, td)
	bottomLeft := intersectLines(bp, bd, lp, ld)
	leftRight := upperLeft.Distance(upperRight)
	topBottom := upperLeft.Distance(bottomLeft)
	return leftRight * topBottom
}

func convexHullAndRotatingCalipers(points []Vector) float64 {
	// ensure 2D
	for i := range points {
		points[i].Z = 0
	}

	// sort points
	sort.Slice(points, func(i, j int) bool {
		if points[i].X == points[j].X {
			return points[i].Y < points[j].Y
		}
		return points[i].X < points[j].X
	})

	// find upper and lower parts of convex hull
	var U, L []Vector
	const eps = 1e-9
	for _, p := range points {
		for len(U) > 1 && cross2D(U[len(U)-2], U[len(U)-1], p) > -eps {
			U = U[:len(U)-1]
		}
		for len(L) > 1 && cross2D(L[len(L)-2], L[len(L)-1], p) < eps {
			L = L[:len(L)-1]
		}
		U = append(U, p)
		L = append(L, p)
	}

	// reverse U
	for i, j := 0, len(U)-1; i < j; i, j = i+1, j-1 {
		U[i], U[j] = U[j], U[i]
	}

	// concat L & U for entire hull
	hull := append(L[1:], U[1:]...)

	// find extrema for initial caliper placement
	leftDir := Vector{0, -1, 0}
	rightDir := Vector{0, 1, 0}
	topDir := Vector{-1, 0, 0}
	bottomDir := Vector{1, 0, 0}
	minPoint := hull[0]
	maxPoint := hull[0]
	var leftIndex, rightIndex, topIndex, bottomIndex int
	for i, p := range hull {
		if p.X < minPoint.X {
			minPoint.X = p.X
			leftIndex = i
		}
		if p.X > maxPoint.X {
			maxPoint.X = p.X
			rightIndex = i
		}
		if p.Y < minPoint.Y {
			minPoint.Y = p.Y
			bottomIndex = i
		}
		if p.Y > maxPoint.Y {
			maxPoint.Y = p.Y
			topIndex = i
		}
	}

	// precompute edge directions
	edgeDirs := make([]Vector, len(hull))
	for i, p := range hull {
		q := hull[(i+1)%len(hull)]
		edgeDirs[i] = q.Sub(p).Normalize()
	}

	// rotating calipers algorithm
	var bestArea, bestResult float64
	for i := range hull {
		leftAngle := math.Acos(leftDir.Dot(edgeDirs[leftIndex]))
		rightAngle := math.Acos(rightDir.Dot(edgeDirs[rightIndex]))
		topAngle := math.Acos(topDir.Dot(edgeDirs[topIndex]))
		bottomAngle := math.Acos(bottomDir.Dot(edgeDirs[bottomIndex]))
		bestAngle := math.Min(leftAngle, math.Min(rightAngle, math.Min(topAngle, bottomAngle)))
		switch {
		case bestAngle == leftAngle:
			leftDir = edgeDirs[leftIndex]
			rightDir = leftDir.Negate()
			bottomDir = leftDir.Perpendicular()
			topDir = bottomDir.Negate()
			leftIndex = (leftIndex + 1) % len(hull)
		case bestAngle == rightAngle:
			rightDir = edgeDirs[rightIndex]
			leftDir = rightDir.Negate()
			bottomDir = leftDir.Perpendicular()
			topDir = bottomDir.Negate()
			rightIndex = (rightIndex + 1) % len(hull)
		case bestAngle == topAngle:
			topDir = edgeDirs[topIndex]
			bottomDir = topDir.Negate()
			rightDir = bottomDir.Perpendicular()
			leftDir = rightDir.Negate()
			topIndex = (topIndex + 1) % len(hull)
		case bestAngle == bottomAngle:
			bottomDir = edgeDirs[bottomIndex]
			topDir = bottomDir.Negate()
			rightDir = bottomDir.Perpendicular()
			leftDir = rightDir.Negate()
			bottomIndex = (bottomIndex + 1) % len(hull)
		}
		area := orientedBoxArea(
			hull[leftIndex], leftDir, hull[rightIndex], rightDir,
			hull[topIndex], topDir, hull[bottomIndex], bottomDir)
		if i == 0 || area < bestArea {
			bestArea = area
			bestResult = math.Atan2(leftDir.Y, leftDir.X)
		}
	}
	return bestResult
}
