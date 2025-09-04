package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var sqlServerDns = "sqlserver://sa:123456@localhost?database=a0001&fetchSize=10000&encrypt=disable"
var pgDsn = "postgres://postgres:123456@localhost:5432/a001?sslmode=disable"
var mySqlDsn = "root:123456@tcp(127.0.0.1:3306)/a001"

func TestOpenDb(t *testing.T) {
	db, err := Open("mysql", mySqlDsn)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	err = db.Ping()
	assert.NoError(t, err)

}
func TestOpenDbMssql(t *testing.T) {
	db, err := Open("sqlserver", sqlServerDns)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	err = db.Ping()
	assert.NoError(t, err)

}
func TestOpenDbPostgres(t *testing.T) {
	db, err := Open("postgres", pgDsn)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	err = db.Ping()
	assert.NoError(t, err)

}
