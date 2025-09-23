package model

import (
	"reflect"
	"sync"

	"github.com/vn-go/dx/internal"
)

type initGetTableName struct {
	once sync.Once
	val  string
	err  error
}

var cacheGetTableName sync.Map

func (reg *modelRegister) getTableName(typ reflect.Type) (string, error) {
	actual, _ := cacheGetTableName.LoadOrStore(typ, &initGetTableName{})
	init := actual.(*initGetTableName)
	init.once.Do(func() {
		init.val, init.err = reg.getTableNameNoCache(typ)
	})
	return init.val, init.err
}
func (reg *modelRegister) getTableNameNoCache(typ reflect.Type) (string, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	// scan field
	typPtr := typ
	if typPtr.Kind() == reflect.Struct {
		typPtr = reflect.PointerTo(typPtr)
	}

	for i := 0; i < typPtr.NumMethod(); i++ {
		if typPtr.Method(i).Name == "Table" && typPtr.Method(i).Type.NumIn() == 1 && typPtr.Method(i).Type.NumOut() == 1 && typPtr.Method(i).Type.Out(0) == reflect.TypeFor[string]() {
			ret := typPtr.Method(i).Func.Call([]reflect.Value{reflect.New(typPtr.Elem())})
			return ret[0].String(), nil
		}
	}
	ret := internal.Utils.SnakeCase(typ.Name())
	ret = internal.Utils.Pluralize(ret)
	return ret, nil
	//return "", fmt.Errorf("model %s has no table tag", typ.String())
}
