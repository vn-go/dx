package dx

type UpdateResult struct {
	RowsAffected int64
	Error        error
	Sql          string //<-- if error is not nil, this field will be not empty
}
