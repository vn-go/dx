package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
)

var pgDsn = "postgres://postgres:123456@localhost:5432/a001?sslmode=disable"
var mySqlDsn = "root:123456@tcp(127.0.0.1:3306)/hrm"

type App struct {
	Name        string    `db:"pk"`
	ShareSecret string    `db:"size:500"`
	CreatedOn   time.Time `db:"ix;default:now()"`
	ModifiedOn  *time.Time
}

func (r *App) Table() string {
	return "sys_apps"
}
func TestMigrateMysql(t *testing.T) {
	dx.AddModels(&App{})

	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		fmt.Println(err.Error())
	}
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
