package utils

import (
	"iter"
	"math/rand"
	"slices"
)

func Append[T any](slices ...[]T) (result []T) {
	var totalLen int

	for _, s := range slices {
		totalLen += len(s)
	}
	result = make([]T, totalLen)

	var i int
	for _, s := range slices {
		i += copy(result[i:], s)
	}

	return
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

func Shuffle[T any](slice []T) (result []T) {
	result = slices.Clone(slice)
	rand.Shuffle(len(slice), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return
}

func Iter[T any](seq iter.Seq[T]) (result []T) {
	result = make([]T, 0)
	for v := range seq {
		result = append(result, v)
	}
	return
}

func Map[T, E any](slice []T, fn func(T) E) (results []E) {
	results = make([]E, len(slice))
	for i, t := range slice {
		results[i] = fn(t)
	}
	return
}

func MapWithErr[T, E any](slice []T, fn func(T) (result E, err error)) (results []E, err error) {
	results = make([]E, len(slice))
	for i, t := range slice {
		results[i], err = fn(t)
		if err != nil {
			return
		}
	}
	return
}

func Maps[T, E any](slice []T, fn func(T) []E) (results []E) {
	results = make([]E, 0, len(slice))
	for _, t := range slice {
		results = append(results, fn(t)...)
	}
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
