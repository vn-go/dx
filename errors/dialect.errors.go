package errors

import (
	"fmt"
	"strings"
)

type DIALECT_DB_ERROR_TYPE = int

const (
	ERR_UNKNOWN DIALECT_DB_ERROR_TYPE = iota
	ERR_DUPLICATE
	ERR_REFERENCES // ✅ refrences_violation
	ERR_REQUIRED
)

func ErrorMessage(t DIALECT_DB_ERROR_TYPE) string {
	switch t {
	case ERR_UNKNOWN:
		return "unknown"
	case ERR_DUPLICATE:
		return "duplicate"
	case ERR_REFERENCES:
		return "references"
	case ERR_REQUIRED:
		return "required"
	default:
		return "unknown"
	}
}
func (e DbError) Error() string {
	return fmt.Sprintf("code=%s, %s: %s cols %v tables %v, entity fields %v", e.Code, ErrorMessage(e.ErrorType), e.ErrorMessage, strings.Join(e.DbCols, ","), e.Table, strings.Join(e.Fields, ","))
}
func (e DbError) Unwrap() error {
	return e.Err
}

type DbError struct {
	Err            error
	ErrorType      DIALECT_DB_ERROR_TYPE
	Code           string
	ErrorMessage   string
	DbCols         []string
	Fields         []string
	Table          string
	StructName     string
	RefTable       string   //<-- table cause error
	RefStructName  string   //<-- Struct cause error
	RefCols        []string //<-- Columns in database cause error
	RefFields      []string //<-- Fields in struct cause error
	ConstraintName string   //<-- Constraint name cause error
}

func (e *DbError) Reload() {
	e.Code = "ERR" + fmt.Sprintf("%04d", e.ErrorType)
	e.ErrorMessage = ErrorMessage(e.ErrorType)
}
