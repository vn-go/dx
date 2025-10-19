package quicky

import "fmt"

type ParseError struct {
	Message         string
	OriginalMessage string
	Args            []interface{}
}

func (e *ParseError) Error() string {
	return e.Message
}
func newParseError(message string, args ...interface{}) error {
	return &ParseError{
		Message:         fmt.Sprintf(message, args...),
		OriginalMessage: message,
		Args:            args,
	}
}
