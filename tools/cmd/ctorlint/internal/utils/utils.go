package utils

func NoError(err error) {
	if err != nil {
		panic(err)
	}
}

func NoErrorT[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
