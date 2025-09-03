package errors

import "fmt"

type MigrationError struct {
	Err error
	Sql string
}

func (e MigrationError) Error() string {
	return fmt.Sprintf("Error executing migration: %s\n%s", e.Err, e.Sql)
}

func CreateError(sql string, err error) MigrationError {
	return MigrationError{
		Err: err,
		Sql: sql,
	}

}
