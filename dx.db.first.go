package dx

import "fmt"

/*
Get first item by filter
@entity
@fiter
@args

	Example:
			db.First(&model,"id={1}",1)
*/
func (db *DB) First(entity interface{}, args ...interface{}) error {
	if len(args) == 0 {
		return db.firstWithNoFilterV2(entity, nil, nil)
	} else if len(args) >= 2 {
		if filter, ok := args[0].(string); ok {
			return db.firstWithFilterV2(entity, filter, nil, nil, args[1:]...)
		} else {
			return fmt.Errorf("first with filter: filter must be string")
		}

	} else {
		return fmt.Errorf("first with filter: filter must be string")
	}
}
