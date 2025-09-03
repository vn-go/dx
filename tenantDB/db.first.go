package tenantDB

// func (db *TenantDB) First(entity interface{}, args ...interface{}) error {
// 	if len(args) == 0 {
// 		return db.firstWithNoFilter(entity)
// 	} else if len(args) >= 2 {
// 		if filter, ok := args[0].(string); ok {
// 			return db.firstWithFilter(entity, filter, args[1:]...)
// 		} else {
// 			return fmt.Errorf("first with filter: filter must be string")
// 		}

// 	} else {
// 		return fmt.Errorf("first with filter: filter must be string")
// 	}
// }
