package errors

type UnsupportedModelError struct {
}

func (e UnsupportedModelError) Error() string {
	return "unsupported model type"
}
