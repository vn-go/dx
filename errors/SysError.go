package errors

type SysError struct {
	message string
}

func (s *SysError) Error() string {
	return s.message
}
func NewSysError(msg string) error {
	return &SysError{
		message: msg,
	}
}
