package errors

import (
	"fmt"
	"strings"
)

type DB_ERR = int

const (
	ERR_UNKNOWN DB_ERR = iota
	ERR_DUPLICATE
	ERR_REFERENCES // ✅ refrences_violation
	ERR_REQUIRED
	ERR_SYSTEM
	ERR_NOT_FOUND
)

func ErrorMessage(t DB_ERR) string {
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
func (e DbErr) Error() string {
	return fmt.Sprintf("code=%s, %s: %s cols %v tables %v, entity fields %v", e.Code, ErrorMessage(e.ErrorType), e.ErrorMessage, strings.Join(e.DbCols, ","), e.Table, strings.Join(e.Fields, ","))
}
func (e DbErr) Unwrap() error {
	return e.Err
}

type DbErr struct {
	Err            error
	ErrorType      DB_ERR
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

func (e *DbErr) Reload() {
	e.Code = "ERR" + fmt.Sprintf("%04d", e.ErrorType)
	e.ErrorMessage = ErrorMessage(e.ErrorType)
}
