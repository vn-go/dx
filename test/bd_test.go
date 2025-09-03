package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

var sqlServerDns = "sqlserver://sa:123456@localhost?database=a0001&fetchSize=10000&encrypt=disable"
var pgDsn = "postgres://postgres:123456@localhost:5432/a001?sslmode=disable"
var mySqlDsn = "root:123456@tcp(127.0.0.1:3306)/a001"

func TestMssqlTenantDB(t *testing.T) {

	db, err := dx.Open("sqlserver", sqlServerDns)

	assert.NoError(t, err)
	defer db.Close()
	err = db.Detect()
	assert.NoError(t, err)
	// do something with the db
}
func TestPgTenantDB(t *testing.T) {

	db, err := dx.Open("postgres", pgDsn)

	assert.NoError(t, err)
	defer db.Close()
	err = db.Detect()
	assert.NoError(t, err)
	// do something with the db
}
func TestMySqlTenantDB(t *testing.T) {

	db, err := dx.Open("mysql", mySqlDsn)

	assert.NoError(t, err)
	defer db.Close()
	err = db.Detect()
	assert.NoError(t, err)
	// do something with the db
}
