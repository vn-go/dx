package test

import (
	"reflect"
	"testing"

	migrate "github.com/vn-go/dx/migrate"
	"github.com/vn-go/dx/migrate/mssql"
	"github.com/vn-go/dx/tenantDB"

	"github.com/stretchr/testify/assert"
)

func Test_mssql_loader(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=a0001&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)

	assert.NoError(t, err)
	err = db.Ping()
	assert.NoError(t, err)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	tables, err := migrator.GetSqlCreateTable(reflect.TypeOf(User{}))
	assert.NoError(t, err)
	assert.NotEmpty(t, tables)

}
func TestLoadFK(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=a0001&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)
	assert.NoError(t, err)

	loader := &mssql.MigratorLoaderMssql{}
	lst, err := loader.LoadForeignKey(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, lst)

}
