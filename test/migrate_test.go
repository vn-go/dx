package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
)

var sqlServerDns = "sqlserver://sa:123456@localhost?database=a001&fetchSize=10000&encrypt=disable"
var pgDsn = "postgres://postgres:123456@localhost:5432/a001?sslmode=disable"
var mySqlDsn = "root:123456@tcp(127.0.0.1:3306)/a001"

func TestMigrateMysql(t *testing.T) {

	db, err := dx.Open("mysql", mySqlDsn)
	dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	assert.NotEmpty(t, db)
	sqls, err := dx.Migrator.GetSql(db.DB)
	assert.NoError(t, err)
	assert.NotEmpty(t, sqls)
}
func TestMigrateMssql(t *testing.T) {
	db, err := dx.Open("sqlserver", sqlServerDns)
	dx.SetManagerDb("sqlserver", "a001")
	assert.NoError(t, err)
	assert.NotEmpty(t, db)
	sqls, err := dx.Migrator.GetSql(db.DB)
	assert.NoError(t, err)
	assert.NotEmpty(t, sqls)
}
func TestMigratePostgres(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	dx.SetManagerDb("postgres", "a001")
	assert.NoError(t, err)
	assert.NotEmpty(t, db)
	sqls, err := dx.Migrator.GetSql(db.DB)
	assert.NoError(t, err)
	assert.NotEmpty(t, sqls)
}
