package utils

func MapSlice[T any, M any](a []T, f func(T) M) []M {
	n := make([]M, len(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

// Filter returns a new slice containing only elements for which f returns true.
func Filter[T any](list []T, f func(T) bool) []T {
	var result []T
	for _, v := range list {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}
