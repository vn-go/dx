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
	Value any
}
type SqlArgs []SqlArg

func (a *SqlArgs) ExtractArgs(args ...any) []any {
	panic("SqlArg ExtractArgs")
}
func (c *SqlArgs) ToSelectorArgs(args []any) SelectorTypesArgs {
	if len(*c) == 0 {
		return SelectorTypesArgs{}
	}
	panic("SqlArgs ToSelectorArgs")
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

//	func (q *ArgOfSqlStruct) GetFormArgs(args []internal.Sq) []SqlArg {
//		ret := []SqlArg{}
//		for _, x := range args {
//			if v, ok := x.(SqlArg); ok {
//				ret = append(ret, v)
//			}
//		}
//		return ret
//	}
// func (q *ArgOfSqlStruct) ExtracAllDynamicArgsFromSelectorTypesArgs(argsSelector SelectorTypesArgs) {
// 	q = FillArrayToEmptyFields[*ArgOfSqlStruct, SqlArg](q)
// 	argsSelectorVal := reflect.ValueOf(argsSelector)
// 	qValue := reflect.ValueOf(*q)
// 	typ := reflect.TypeFor[SelectorTypesArgs]()
// 	typOfDynamicArg := reflect.TypeFor[ArgOfSqlStruct]()
// 	//typOfDynamicArgs := reflect.TypeFor[[]DynamicArg]()
// 	for i := 0; i < typ.NumField(); i++ {
// 		f := typ.Field(i)
// 		argsSelectorValField := argsSelectorVal.FieldByIndex(f.Index)
// 		arr := qValue.FieldByIndex(f.Index)
// 		// if arr.IsNil() {
// 		// 	fmt.Println(f.Name)
// 		// 	arr.Set(reflect.MakeSlice(arr.Elem().Type(), 0, 0))
// 		// }

// 		for j := 0; j < argsSelectorValField.Len(); j++ {
// 			fmt.Println(argsSelectorValField.Index(i).Elem().Type().String())
// 			if argsSelectorValField.Index(j).Elem().Type() == typOfDynamicArg {
// 				arr = reflect.Append(arr, argsSelectorValField.Index(j).Elem())
// 			}
// 		}
// 		// if argsSelectorValField.IsValid() && qValue.FieldByIndex(f.Index).IsValid() {

// 		// }
// 	}
// }
// func (q *ArgOfSqlStruct) BuildDyanmicArgsToSelectorTypesArgs(args []any) SelectorTypesArgs {
// 	q = FillArrayToEmptyFields[*ArgOfSqlStruct, SqlArg](q)
// 	ret := SelectorTypesArgs{
// 		ArgWhere:   []SqlArg{},
// 		ArgsSelect: []SqlArg{},
// 		ArgJoin:    []SqlArg{},
// 		ArgGroup:   []SqlArg{},
// 		ArgHaving:  []SqlArg{},
// 		ArgOrder:   []SqlArg{},
// 		ArgSetter:  []SqlArg{},
// 	}
// 	q = FillArrayToEmptyFields[*ArgOfSqlStruct, SqlArg](q)
// 	valOfRet := reflect.ValueOf(&ret).Elem() // dùng Elem() để set được field
// 	val := reflect.ValueOf(*q)
// 	ft := reflect.TypeFor[SqlArg]()

// 	for i := 0; i < ft.NumField(); i++ {
// 		f := ft.Field(i)

// 		if f.Name == "ArgsSelect" {
// 			fmt.Println("collect args :" + f.Name)
// 		}
// 		fv := val.FieldByIndex(f.Index)
// 		if !fv.IsValid() {
// 			continue
// 		}
// 		if fv.IsNil() {

// 			fv.Set(reflect.MakeSlice(reflect.TypeOf([]SqlArg{}), 0, 0).Addr())
// 		}
// 		qas := fv.Interface().([]SqlArg)
// 		arr := reflect.MakeSlice(reflect.TypeOf([]any{}), 0, 0)

// 		for _, qa := range qas {
// 			if qa.Index >= 1 && qa.Index <= len(args) {
// 				arr = reflect.Append(arr, reflect.ValueOf(args[qa.Index-1]))
// 			}
// 		}

// 		fs := valOfRet.FieldByIndex(f.Index)
// 		if fs.IsValid() && fs.CanSet() {
// 			fs.Set(arr)
// 		}
// 	}

// 	return ret
// }

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
					fmt.Println(x)

				}
				valueField.Elem().Set(argsValue) // panic cho nay
			}
		}
	}
	return ret
}
func (compilerArgs *CompilerArgs) ToSelectorArgs(args []any) SelectorTypesArgs {
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
			if x.IsDynamic {
				argsValue = reflect.Append(argsValue, reflect.ValueOf(args[x.Index]))
			} else {
				argsValue = reflect.Append(argsValue, reflect.ValueOf(x.Value))
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
