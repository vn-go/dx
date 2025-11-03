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

	switch {
	case t.AssignableTo(reflect.TypeOf(sql.NullInt64{})):
		return reflect.TypeOf(sql.NullInt64{})
	case t.AssignableTo(reflect.TypeOf(sql.NullFloat64{})):
		return reflect.TypeOf(sql.NullFloat64{})
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.TypeOf(sql.NullInt64{})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.TypeOf(sql.NullInt64{})
	case reflect.Float32, reflect.Float64:
		return reflect.TypeOf(sql.NullFloat64{})
	default:
		return nil
	}
}

// --- AVG ------------------------------------------------------

func SQLFuncAvgReturnType(argTypes []reflect.Type) reflect.Type {
	if len(argTypes) == 0 {
		return nil
	}
	t := unwrapType(argTypes[0])

	switch {
	case t.AssignableTo(reflect.TypeOf(sql.NullFloat64{})):
		return reflect.TypeOf(sql.NullFloat64{})
	}

	// AVG luôn trả về float64 hoặc sql.NullFloat64
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return reflect.TypeOf(sql.NullFloat64{})
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
		return reflect.TypeOf(sql.NullInt64{})
	case t.AssignableTo(reflect.TypeOf(sql.NullFloat64{})):
		return reflect.TypeOf(sql.NullFloat64{})
	case t.AssignableTo(reflect.TypeOf(sql.NullString{})):
		return reflect.TypeOf(sql.NullString{})
	case t.AssignableTo(reflect.TypeOf(sql.NullTime{})):
		return reflect.TypeOf(sql.NullTime{})
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.TypeOf(sql.NullInt64{})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.TypeOf(sql.NullInt64{})
	case reflect.Float32, reflect.Float64:
		return reflect.TypeOf(sql.NullFloat64{})
	case reflect.String:
		return reflect.TypeOf(sql.NullString{})
	case reflect.Struct:
		if t.AssignableTo(reflect.TypeOf(time.Time{})) {
			return reflect.TypeOf(sql.NullTime{})
		}
	}
	return nil
}

// --- COUNT ----------------------------------------------------

func SQLFuncCountReturnType(_ []reflect.Type) reflect.Type {
	// COUNT luôn trả về số nguyên
	return reflect.TypeOf(sql.NullInt64{})
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

	panic(fmt.Sprintf("Unsupported function: %s, ref helperType.CombineTypeByFunc,file %s", funcName, `internal\helper.CombineTypeByFunc.go`))
}
func (c *helperType) CombineDbTypeByFunc(funcName string, argTypes []reflect.Type) sqlparser.ValType {
	return c.GetSqlTypeFfromGoType(c.CombineTypeByFunc(funcName, argTypes))
}
