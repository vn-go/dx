package dx

import "reflect"

func (db *DB) firstWithFilter(entity interface{}, filter string, args ...interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, err := db.buildBasicSqlFirstItem(typ, filter) //OnBuildSQLFirstItem(typ, db, filter)
	if err != nil {
		return err
	}
	return db.ExecToItem(entity, sql, args...)

}
