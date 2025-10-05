package dx

import (
	"reflect"

	"github.com/vn-go/dx/internal"
)

func (db *DB) Joins(query string, args ...interface{}) *selectorTypes {
	return &selectorTypes{
		db:      db,
		strJoin: query,
		args: internal.SelectorTypesArgs{
			ArgJoin: args,
		},
	}

}
func (db *DB) From(model any) *selectorTypes {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return &selectorTypes{
		db:         db,
		entityType: &typ,
	}
}
