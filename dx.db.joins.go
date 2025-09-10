package dx

func (db *DB) Joins(query string, args ...interface{}) *selectorTypes {
	return &selectorTypes{
		db:      db,
		strJoin: query,
		args: selectorTypesArgs{
			ArgJoin: args,
		},
	}

}
