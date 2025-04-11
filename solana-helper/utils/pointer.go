package utils

func Pointer[T any](v T) *T {
	return &v
}

func Default[T any](v *T, def T) T {
	if v == nil {
		return def
	} else {
		return *v
	}
}

func IgnoreErr[T any](v T, err error) T {
	return v
}
