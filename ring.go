package fauxgl

import "math"

type Ring []Vector

func (r Ring) At(i int) Vector {
	n := len(r)
	if i < 0 {
		i += n
	}
	return r[i%n]
}

func (r Ring) Length() float64 {
	var sum float64
	for i := range r {
		sum += r.At(i).Distance(r.At(i + 1))
	}
	return sum
}

func (r Ring) Simplify() Ring {
	n := len(r)
	straight := make([]bool, n)
	start := -1
	for i := 0; i < n; i++ {
		p1 := r.At(i - 1)
		p2 := r.At(i)
		p3 := r.At(i + 1)
		v1 := p2.Sub(p1).Normalize()
		v2 := p3.Sub(p2).Normalize()
		c := v1.Dot(v2)
		straight[i] = math.Abs(c-1) < 1e-9
		if start < 0 && !straight[i] {
			start = i
		}
	}
	var result []Vector
	for i := start; i < start+n; i++ {
		j := i % n
		if !straight[j] {
			result = append(result, r[j])
		}
	}
	return result
}
