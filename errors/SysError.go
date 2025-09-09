package errors

type SysError struct {
	message string
}

func (s *SysError) Error() string {
	return s.message
}
func NewSysError(msg string) error {
	return &SysError{
		message: msg,
	}
}

type SqlExecError struct {
	message  string
	sql      string
	dbDriver string
	err      error
}

func (s *SqlExecError) Error() string {
	return s.message + "\nsql\t:" + s.sql + "\ndb\t:" + s.dbDriver + "\n" + s.err.Error()
}
func NewSqlExecError(msg, sql, dbDriver string, err error) error {
	return &SqlExecError{
		message:  msg,
		sql:      sql,
		dbDriver: dbDriver,
		err:      err,
	}
}
