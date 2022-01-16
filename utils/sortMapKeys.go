package utils

import "sort"

func SortedMapKeys[T any](input map[string]T) []string {
	keys := make([]string, len(input))
	i := 0
	for k := range input {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
