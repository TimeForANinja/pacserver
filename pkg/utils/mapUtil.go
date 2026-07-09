package utils

// MapToArray extracts the values from a map into a slice.
func MapToArray[K comparable, V any](m map[K]V) []V {
	var arr []V
	for _, k := range m {
		arr = append(arr, k)
	}
	return arr
}

// MapClone copies a map into a new allocation.
func MapClone[K comparable, V any](m map[K]V) map[K]V {
	var clone = make(map[K]V)
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// MergeMaps overlays the second map on top of the first one.
func MergeMaps[K comparable, V any](m1 map[K]V, m2 map[K]V) map[K]V {
	var clone = make(map[K]V)
	for k, v := range m1 {
		clone[k] = v
	}
	for k, v := range m2 {
		clone[k] = v
	}
	return clone
}

// MapFromArray builds a keyed map from a slice and key extractor.
func MapFromArray[K comparable, V any](arr []V, get_key func(V) K) map[K]V {
	var m = make(map[K]V)
	if get_key == nil {
		return m
	}
	for _, v := range arr {
		m[get_key(v)] = v
	}
	return m
}
