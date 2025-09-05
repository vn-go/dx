package errors

type NotFoundErr struct {
}

func (e *NotFoundErr) Error() string {
	return "Not Found"
}
func NewNotFoundErr() error {
	return &NotFoundErr{}
}
