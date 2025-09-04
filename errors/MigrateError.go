package errors

import "fmt"

type MigrationError struct {
	Err error
	Sql string
}

func (e MigrationError) Error() string {
	return fmt.Sprintf("Error executing migration: %s\n%s", e.Err, e.Sql)
}

func NewMigrationError(sql string, err error) error {
	return MigrationError{
		Err: err,
		Sql: sql,
	}

}
