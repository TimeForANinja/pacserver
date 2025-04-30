package utils

func IfIsNil[T comparable](val *T, def T) T {
	if val == nil {
		return def
	}
	return *val
}
