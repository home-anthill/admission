package utils

import "slices"

func MapSlice[T any, M any](a []T, f func(T) M) []M {
	n := make([]M, len(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

// https://stackoverflow.com/a/78185810/3590376
func Filter[T any, M bool](list []T, f func(T) M) []T {
	return slices.Collect(
		func(yield func(T) bool) {
			for _, v := range list {
				if f(v) {
					if !yield(v) {
						return
					}
				}
			}
		},
	)
}
