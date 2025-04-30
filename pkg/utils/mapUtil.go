package utils

func MapToArray[K comparable, V any](m map[K]V) []V {
	var arr []V
	for _, k := range m {
		arr = append(arr, k)
	}
	return arr
}
