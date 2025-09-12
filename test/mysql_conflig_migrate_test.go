package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestMigrateConflictMysql(t *testing.T) {
	dx.SkipLoadSchemOnMigrate(true)
	db, err := dx.Open("mysql", mySqlDsn)
	assert.NoError(t, err)
	defer db.Close()
}
func TestMigrateConflictPG(t *testing.T) {
	dx.SkipLoadSchemOnMigrate(true)
	db, err := dx.Open("postgres", pgDsn)
	assert.NoError(t, err)
	defer db.Close()
}
func TestMigrateConflictMssql(t *testing.T) {
	dx.SkipLoadSchemOnMigrate(true)
	db, err := dx.Open("sqlserver", sqlServerDns)
	assert.NoError(t, err)
	defer db.Close()
}
