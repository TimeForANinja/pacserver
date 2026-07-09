package utils

func MapToArray[K comparable, V any](m map[K]V) []V {
	var arr []V
	for _, k := range m {
		arr = append(arr, k)
	}
	return arr
}

func MapClone[K comparable, V any](m map[K]V) map[K]V {
	var clone = make(map[K]V)
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

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

func MapFromArray[K comparable, V any](arr []V, get_key func(V) K) map[K]V {
	var m = make(map[K]V)
	for _, v := range arr {
		m[get_key(v)] = v
	}
	return m
}
