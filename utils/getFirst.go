package utils

func GetFirst[T any](opts []T) *T {
	if len(opts) == 0 {
		return nil
	}
	return &opts[0]
}
