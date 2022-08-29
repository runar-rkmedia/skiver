package utils

// unwraps the first value, panics if err
func Must[T any](t T, err error) T {

	if err != nil {
		panic(err)
	}
	return t

}
