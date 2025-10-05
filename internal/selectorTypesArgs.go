package internal

import (
	"reflect"
	"sync"
)

type QuestionArg struct {
	Index int
}
type SelectorTypesArgs struct {
	ArgWhere   []any
	ArgsSelect []any
	ArgJoin    []any
	ArgGroup   []any
	ArgHaving  []any
	ArgOrder   []any
	ArgSetter  []any
}

func (a *SelectorTypesArgs) GetArgs(fields []reflect.StructField) []any {
	ret := []any{}
	val := reflect.ValueOf(*a)
	for _, f := range fields {
		fv := val.FieldByIndex(f.Index)
		if fv.IsValid() {
			//"reflect.Value.Elem"
			if fv.IsNil() {
				continue
			}

			ret = append(ret, fv.Interface().([]any)...)
		}

	}
	return ret
}

type SqlInfoArgs struct {
	ArgWhere   reflect.StructField
	ArgsSelect reflect.StructField
	ArgJoin    reflect.StructField
	ArgGroup   reflect.StructField
	ArgHaving  reflect.StructField
	ArgOrder   reflect.StructField
	ArgSetter  reflect.StructField
}

var selectorTypesArgsGetFields = &SqlInfoArgs{}
var selectorTypesArgsGetFieldsOnce sync.Once

func (a *SelectorTypesArgs) GetFields() *SqlInfoArgs {
	selectorTypesArgsGetFieldsOnce.Do(func() {
		v := reflect.ValueOf(selectorTypesArgsGetFields).Elem()
		typ := reflect.TypeFor[SelectorTypesArgs]()
		for i := 0; i < typ.NumField(); i++ {
			vf := v.FieldByName(typ.Field(i).Name)
			if vf.IsValid() {
				vf.Set(reflect.ValueOf(typ.Field(i)))
			}
		}

	})
	return selectorTypesArgsGetFields
}
