package dx

func (db *DB) Joins(query string, args ...interface{}) *selectorTypes {
	return &selectorTypes{
		db:      db,
		strJoin: query,
		argJoin: args,
	}

}
