package main

func noerror(err error) {
	if err != nil {
		panic(err)
	}
}

func noerrorT[T interface{}](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
