package internal

import (
	"fmt"
	"reflect"
	"sync"
)

//	type DynamicArg struct {
//		Index int
//	}
type SqlArg struct {
	// Exmaple select a+1,b+?=> first is static the second is dynamic
	IsDynamic bool
	//index in sql
	Index int
	// if not is dynamic this value will part from sql exmaple select a+1,b+? index 0 is
	Value        any
	IsInTextArgs bool
	TextArgIndex int
}
type SqlArgs []SqlArg

func (a *SqlArgs) ExtractArgs(args ...any) []any {
	ret := []any{}
	for _, x := range *a {
		if x.IsDynamic {
			if x.Index < len(args) {
				ret = append(ret, args[x.Index])
			} else {
				ret = append(ret, x.Value)
			}
		}
	}
	//panic("SqlArg ExtractArgs")
	return ret
}
func (c *SqlArgs) ToSelectorArgs(args []any) SelectorTypesArgs {
	if len(*c) == 0 {
		return SelectorTypesArgs{}
	}
	panic("SqlArgs ToSelectorArgs")
}
func UnionCompilerArgs(a CompilerArgs, b CompilerArgs) CompilerArgs {
	ret := CompilerArgs{
		ArgWhere:   append(a.ArgWhere, b.ArgWhere...),
		ArgsSelect: append(a.ArgsSelect, b.ArgsSelect...),
		ArgJoin:    append(a.ArgJoin, b.ArgJoin...),
		ArgGroup:   append(a.ArgGroup, b.ArgGroup...),
		ArgHaving:  append(a.ArgHaving, b.ArgHaving...),
		ArgOrder:   append(a.ArgOrder, b.ArgOrder...),
		ArgSetter:  append(a.ArgSetter, b.ArgSetter...),
	}
	return ret

}
func FillArrayToEmptyFields[TObj any, TField any](obj TObj) TObj {
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	// Trường hợp obj là pointer tới struct
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i).Type
		if fieldType != reflect.TypeOf([]TField{}) {
			continue
		}

		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Slice && fieldValue.IsNil() {
			fieldValue.Set(reflect.MakeSlice(fieldType, 0, 0))
		}
	}

	return obj
}
func GetAllElemetsByType[TType any](args []any) []TType {
	ret := []TType{}
	for _, x := range args {
		if r, ok := x.(TType); ok {
			ret = append(ret, r)
		}
	}
	return ret
}

type CompilerArgs struct {
	ArgWhere   SqlArgs
	ArgsSelect SqlArgs
	ArgJoin    SqlArgs
	ArgGroup   SqlArgs
	ArgHaving  SqlArgs
	ArgOrder   SqlArgs
	ArgSetter  SqlArgs
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

func NewSelectorTypesArgs() SelectorTypesArgs {
	return SelectorTypesArgs{
		ArgWhere:   []any{},
		ArgsSelect: []any{},
		ArgJoin:    []any{},
		ArgGroup:   []any{},
		ArgHaving:  []any{},
		ArgOrder:   []any{},
		ArgSetter:  []any{},
	}
}
func (compilerArgs *CompilerArgs) ToSelectorArgs1(args []any) SelectorTypesArgs {
	typ := reflect.TypeFor[CompilerArgs]()

	ret := NewSelectorTypesArgs()
	valueOfCompilerArgs := reflect.ValueOf(*compilerArgs)
	retValue := reflect.ValueOf(ret)
	for i := 0; i < typ.NumField(); i++ {
		valueOfField := valueOfCompilerArgs.FieldByIndex(typ.Field(i).Index)
		if valueOfField.IsValid() {
			if !valueOfField.IsNil() {
				valueField := retValue.FieldByIndex(typ.Field(i).Index)
				items := valueOfField.Interface().(SqlArgs)
				argsValue := reflect.MakeSlice(reflect.TypeFor[[]any](), 0, 0)
				for _, x := range items {
					if x.IsDynamic {
						argsValue = reflect.Append(argsValue, reflect.ValueOf(args[x.Index]))
						//args = append(args, reflect.ValueOf(args[x.Index]))
					} else {
						argsValue = reflect.Append(argsValue, reflect.ValueOf(x.Value))
						//args = append(args, reflect.ValueOf(x.Value))
					}

				}
				valueField.Elem().Set(argsValue) // panic cho nay
			}
		}
	}
	return ret
}
func (compilerArgs *CompilerArgs) ToSelectorArgs(args []any, extraTextArgs []string) SelectorTypesArgs {
	typ := reflect.TypeFor[CompilerArgs]()
	ret := NewSelectorTypesArgs()

	valueOfCompilerArgs := reflect.ValueOf(*compilerArgs)
	retValue := reflect.ValueOf(&ret).Elem() // make fields settable

	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		valueOfField := valueOfCompilerArgs.FieldByIndex(fieldType.Index)

		if !valueOfField.IsValid() || valueOfField.IsNil() {
			continue
		}

		valueField := retValue.FieldByIndex(fieldType.Index)
		items := valueOfField.Interface().(SqlArgs)
		argsValue := reflect.MakeSlice(reflect.TypeFor[[]any](), 0, 0)

		for _, x := range items {
			if !x.IsInTextArgs {
				if x.IsDynamic {
					argsValue = reflect.Append(argsValue, reflect.ValueOf(args[x.Index]))
				} else {
					argsValue = reflect.Append(argsValue, reflect.ValueOf(x.Value))
				}
			} else {
				argsValue = reflect.Append(argsValue, reflect.ValueOf(extraTextArgs[x.TextArgIndex]))
			}
		}

		// ✅ set slice directly
		valueField.Set(argsValue)
	}

	return ret
}

func (a *SelectorTypesArgs) GetArgs(fields []reflect.StructField) []any {
	ret := []any{}
	val := reflect.ValueOf(*a)
	//var nilField reflect.StructField
	for _, f := range fields {
		if f.Name == "" {
			continue
		}
		fv := val.FieldByIndex(f.Index)

		if fv.IsValid() {
			//"reflect.Value.Elem"
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Struct { //<-- arg of query can not be struct
				fmt.Println(f.Name)
				dataTest := fv.Interface()
				fmt.Println(dataTest)

				//ret = append(ret, fv.Interface())
				continue
			} else {
				ret = append(ret, fv.Interface().([]any)...)
			}

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
