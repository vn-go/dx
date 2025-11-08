package internal

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/vn-go/dx/sqlparser"
)

// --- SUM ------------------------------------------------------

func SQLFuncSumReturnType(argTypes []reflect.Type) reflect.Type {
	if len(argTypes) == 0 {
		return nil
	}
	t := unwrapType(argTypes[0])

	// Nếu là sql.NullInt64 hoặc tương tự -> trả về type Go cơ bản
	switch {
	case t.AssignableTo(reflect.TypeOf(sql.NullInt64{})):
		return reflect.TypeOf(int64(0))
	case t.AssignableTo(reflect.TypeOf(sql.NullFloat64{})):
		return reflect.TypeOf(float64(0))
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// SUM của số nguyên có thể overflow, nên thường trả về int64
		return reflect.TypeOf(int64(0))
	case reflect.Float32, reflect.Float64:
		return reflect.TypeOf(float64(0))
	default:
		return reflect.TypeFor[any]()
	}
}

// --- AVG ------------------------------------------------------

func SQLFuncAvgReturnType(argTypes []reflect.Type) reflect.Type {
	if len(argTypes) == 0 {
		return nil
	}
	t := unwrapType(argTypes[0])

	// Nếu là sql.NullFloat64 hoặc kiểu float -> AVG luôn trả float64
	switch {
	case t.AssignableTo(reflect.TypeOf(sql.NullFloat64{})):
		return reflect.TypeOf(float64(0))
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return reflect.TypeOf(float64(0))
	default:
		return nil
	}
}

// --- MIN & MAX ------------------------------------------------

func SQLFuncMinMaxReturnType(argTypes []reflect.Type) reflect.Type {
	if len(argTypes) == 0 {
		return nil
	}
	t := unwrapType(argTypes[0])

	switch {
	case t.AssignableTo(reflect.TypeOf(sql.NullInt64{})):
		return reflect.TypeOf(int64(0))
	case t.AssignableTo(reflect.TypeOf(sql.NullFloat64{})):
		return reflect.TypeOf(float64(0))
	case t.AssignableTo(reflect.TypeOf(sql.NullString{})):
		return reflect.TypeOf("")
	case t.AssignableTo(reflect.TypeOf(sql.NullTime{})):
		return reflect.TypeOf(time.Time{})
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.TypeOf(int64(0))
	case reflect.Float32, reflect.Float64:
		return reflect.TypeOf(float64(0))
	case reflect.String:
		return reflect.TypeOf("")
	case reflect.Struct:
		if t.AssignableTo(reflect.TypeOf(time.Time{})) {
			return reflect.TypeOf(time.Time{})
		}
	}

	return nil
}

// --- COUNT ----------------------------------------------------

func SQLFuncCountReturnType(_ []reflect.Type) reflect.Type {
	// COUNT trong SQL luôn trả về số nguyên (int64)
	return reflect.TypeOf(int64(0))
}

// --- Helper ---------------------------------------------------

func unwrapType(t reflect.Type) reflect.Type {
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// --- Registry -------------------------------------------------

var SQLAggregateTypeInfer = map[string]func([]reflect.Type) reflect.Type{
	"SUM":   SQLFuncSumReturnType,
	"AVG":   SQLFuncAvgReturnType,
	"MIN":   SQLFuncMinMaxReturnType,
	"MAX":   SQLFuncMinMaxReturnType,
	"COUNT": SQLFuncCountReturnType,
}
var mapFuncReturnTypes = map[string]reflect.Type{
	"CONCAT": reflect.TypeFor[*string](),
}

func (c *helperType) CombineTypeByFunc(funcName string, argTypes []reflect.Type) reflect.Type {
	funcName = strings.ToUpper(funcName)
	if r, ok := SQLAggregateTypeInfer[funcName]; ok {
		return r(argTypes)
	}
	if r, ok := mapFuncReturnTypes[funcName]; ok {
		return r
	}
	if funcName == "IF" {
		return reflect.TypeFor[any]()
	}
	if funcName == "LEN" {
		return reflect.TypeFor[int64]()
	}
	panic(fmt.Sprintf("Unsupported function: %s, ref helperType.CombineTypeByFunc,file %s", funcName, `internal\helper.CombineTypeByFunc.go`))
}
func (c *helperType) CombineDbTypeByFunc(funcName string, argTypes []reflect.Type) sqlparser.ValType {
	return c.GetSqlTypeFfromGoType(c.CombineTypeByFunc(funcName, argTypes))
}
