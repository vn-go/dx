package test

import (
	"fmt"
	"reflect"
	"testing"

	migrate "github.com/vn-go/dx/migrate"
	"github.com/vn-go/dx/migrate/postgres"
	"github.com/vn-go/dx/tenantDB"

	"github.com/stretchr/testify/assert"
)

func TestPGMigrate(t *testing.T) {
	// pg dsn
	pgDsn := "postgres://postgres:123456@localhost:5432/fx001?sslmode=disable"
	// create new migrate instance
	db, err := tenantDB.Open("postgres", pgDsn)

	assert.NoError(t, err)

	migrator, err := migrate.NewMigrator(db)
	loader := migrator.GetLoader()
	pgLoader := loader.(*postgres.MigratorLoaderPostgres)
	cols, err := pgLoader.LoadAllTable(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, cols)
	pks, err := pgLoader.LoadAllPrimaryKey(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, pks)
	uk, err := pgLoader.LoadAllUniIndex(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, uk)
	idx, err := pgLoader.LoadAllIndex(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, idx)
	fk, err := pgLoader.LoadForeignKey(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, fk)
	schema, err := pgLoader.LoadFullSchema(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, schema)
	tables, err := migrator.GetSqlCreateTable(reflect.TypeOf(User{}))
	assert.NoError(t, err)
	assert.NotEmpty(t, tables)
}
func TestPGGenerateSQLCreateTable(t *testing.T) {
	pgDsn := "postgres://postgres:123456@localhost:5432/fx001?sslmode=disable"
	// create new migrate instance
	db, err := tenantDB.Open("postgres", pgDsn)

	assert.NoError(t, err)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	pgm := migrator.(*postgres.MigratorPostgres)
	sql, err := pgm.GetSqlCreateTable(reflect.TypeOf(User{}))
	assert.NoError(t, err)

	fmt.Println(sql)
	assert.NotEmpty(t, sql)

}
func TestPGSqlAddColumns(t *testing.T) {
	pgDsn := "postgres://postgres:123456@localhost:5432/fx001?sslmode=disable"

	// create new migrate instance
	db, err := tenantDB.Open("postgres", pgDsn)

	assert.NoError(t, err)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	pgm := migrator.(*postgres.MigratorPostgres)
	sql, err := pgm.GetSqlAddColumn(reflect.TypeOf(User{}))
	assert.NoError(t, err)

	fmt.Println(sql)
	assert.NotEmpty(t, sql)
}
func TestPGAddIndex(t *testing.T) {
	pgDsn := "postgres://postgres:123456@localhost:5432/fx001?sslmode=disable"

	// create new migrate instance
	db, err := tenantDB.Open("postgres", pgDsn)

	assert.NoError(t, err)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	pgm := migrator.(*postgres.MigratorPostgres)
	sql, err := pgm.GetSqlAddIndex(reflect.TypeOf(User{}))
	assert.NoError(t, err)

	fmt.Println(sql)
	assert.NotEmpty(t, sql)
}
func TestGetSqlAddUniqueIndex(t *testing.T) {
	pgDsn := "postgres://postgres:123456@localhost:5432/fx001?sslmode=disable"

	// create new migrate instance
	db, err := tenantDB.Open("postgres", pgDsn)

	assert.NoError(t, err)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	pgm := migrator.(*postgres.MigratorPostgres)
	sql, err := pgm.GetSqlAddUniqueIndex(reflect.TypeOf(User{}))
	assert.NoError(t, err)

	fmt.Println(sql)
	assert.NotEmpty(t, sql)
}
func TestGetAddForeignKey(t *testing.T) {
	pgDsn := "postgres://postgres:123456@localhost:5432/fx001?sslmode=disable"

	// create new migrate instance
	db, err := tenantDB.Open("postgres", pgDsn)

	assert.NoError(t, err)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	pgm := migrator.(*postgres.MigratorPostgres)
	sql, err := pgm.GetSqlAddForeignKey()
	assert.NoError(t, err)

	fmt.Println(sql)
	assert.NotEmpty(t, sql)
}
func BenchmarkLoadFullSchema(b *testing.B) {
	pgDsn := "postgres://postgres:123456@localhost:5432/a001?sslmode=disable"
	db, err := tenantDB.Open("postgres", pgDsn)
	assert.NoError(b, err)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(b, err)

	loader := migrator.GetLoader()
	pgLoader := loader.(*postgres.MigratorLoaderPostgres)
	b.ResetTimer() // Reset timer để chỉ đo phần bên dưới
	for i := 0; i < b.N; i++ {

		schema, err := pgLoader.LoadFullSchema(db)
		assert.NoError(b, err)
		assert.NotEmpty(b, schema)

	}
}
