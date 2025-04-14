package Utils

func MapFilterCount[K comparable, V any](m map[K]V, condition func(v V) bool) (count int) {
	for _, v := range m {
		if condition(v) {
			count++
		}
	}
	return
}
