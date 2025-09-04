package errors

type UnsupportedError struct {
	message string
}

func (e *UnsupportedError) Error() string {
	return e.message
}
func NewUnsupportedError(msg string) error {
	return &UnsupportedError{
		message: msg,
	}
}
