package lib

func AnyError(callback func(error), functions ...func() error) {
	for _, f := range functions {
		err := f()
		if err != nil {
			callback(err)
			return
		}
	}
}
