package compiler

import "fmt"

type compilerError struct {
	msg     string
	errType ERR_TYPE
}
type ERR_TYPE int

const (
	ERR ERR_TYPE = iota
	ERR_TABLE_NOT_FOUND
)

func (e ERR_TYPE) string() string {
	if e == ERR_TABLE_NOT_FOUND {
		return "table is not found"
	}
	return "compiler sql error"
}
func (c *compilerError) Error() string {
	return fmt.Sprintf("%s %d", c.msg, c.errType)

}
func newCompilerError(msg string, errType ERR_TYPE) error {
	return &compilerError{
		msg:     msg,
		errType: errType,
	}
}
