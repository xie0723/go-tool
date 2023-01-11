package collection

// map 方法
func Map[T any, M any](a []T, f func(T) M) []M {
	n := make([]M, len(a), cap(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

// reduce 方法
func Reduce[T any, M any](a []T, f func(M, T) M, initial M) M {
	if len(a) == 0 || f == nil {
		var vv M
		return vv
	}

	l := len(a) - 1
	reduce := func(a []T, ff func(M, T) M, memo M, startPoint, direction, length int) M {
		result := memo
		index := startPoint
		for i := 0; i <= length; i++ {
			result = ff(result, a[index])
			index += direction
		}
		return result
	}
	return reduce(a, f, initial, 0, 1, l)
}
