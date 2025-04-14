package Utils

import (
	"github.com/samber/lo"
)

func Append[T any](slices ...[]T) (result []T) {
	return lo.Flatten(slices)
}

func Grow[T any](slice []T, capability int) (result []T) {
	result = make([]T, len(slice), max(capability, len(slice), cap(slice)))

	copy(result, slice)

	return
}

func GrowSize[T any](slice []T, capability int) (result []T) {
	result = make([]T, max(capability, len(slice)), max(capability, len(slice), cap(slice)))

	copy(result, slice)

	return
}

func MapsWithErr[T, E any](slice []T, fn func(T) (result []E, err error)) (results []E, err error) {
	results = make([]E, 0, len(slice))
	for _, t := range slice {
		var result []E
		result, err = fn(t)
		if err != nil {
			return
		}
		results = append(results, result...)
	}
	return
}
