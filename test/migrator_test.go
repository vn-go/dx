package test

import (
	"reflect"
	"testing"
	"time"

	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx"
	"github.com/vn-go/dx/tenantDB"

	"github.com/stretchr/testify/assert"
	migrate "github.com/vn-go/dx/migrate"
)

type User struct {
	ID           int       `db:"pk;auto"`                // primary key, auto increment
	Name         string    `db:"pk;size:50;index"`       // mapped column name, varchar(50)
	Email        string    `db:"uk:test_email;size:120"` // unique constraint named "test_email"
	Profile      *string   `db:"size:255"`               // nullable string
	CreatedAt    time.Time `db:"default:now()"`          // default timestamp
	Price        float64   `db:"type:decimal(10,2)"`     // custom type and column name
	HashPassword string    `db:"size(250)"`
	Username     string    `db:"size(50)"`
}
type User2 struct {
	ID    int    `db:"pk;auto"`
	Name  string `db:"pk;size:50;index:name_email_idx"` // mapped column name, varchar(50)
	Email string `db:"uk:test_email;size:120;index:name_email_idx"`
}
type User3 struct {
	Name  string `db:"uk:name_email_idx;size:50"` // mapped column name, varchar(50)
	Email string `db:"uk:name_email_idx;size:120"`
}

func init() {
	dx.ModelRegistry.Add(&User{})
}

func TestMigrator(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=a0001&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	tables, err := migrator.GetSqlCreateTable(reflect.TypeOf(User{}))
	assert.NoError(t, err)

	expected := `CREATE TABLE [users] (
  [id] int IDENTITY(1,1) NOT NULL,
  [name] nvarchar(50) NOT NULL,
  [email] nvarchar(120) NOT NULL,
  [profile] nvarchar(255) NULL,
  [created_at] datetime2 NOT NULL DEFAULT GETDATE(),
  [price] float NOT NULL,
  CONSTRAINT [PK_users__id_name] PRIMARY KEY ([id], [name])
)`
	assert.Equal(t, expected, tables)

	// TODO: implement test cases
}
func TestMigratorIndex(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=aaa&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	script, err := migrator.GetSqlAddIndex(reflect.TypeOf(User{}))
	assert.NoError(t, err)
	assert.Equal(t, "CREATE INDEX [IDX_users__name] ON [users] ([name])", script)
}
func TestMigratorIndex2Cols(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=aaa&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	script, err := migrator.GetSqlAddIndex(reflect.TypeOf(User2{}))
	assert.NoError(t, err)
	assert.Equal(t, "CREATE INDEX [IDX_users2__name_email] ON [users2] ([name], [email])", script)
}
func TestMigratorUniqueIndex(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=aaa&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	script, err := migrator.GetSqlAddUniqueIndex(reflect.TypeOf(User{}))
	assert.NoError(t, err)
	assert.Equal(t, "CREATE INDEX [IDX_users__name] ON [users] ([name])", script)
}
func TestMigratorUniqueIndex2Cols(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=aaa&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	script, err := migrator.GetSqlAddUniqueIndex(reflect.TypeOf(User3{}))
	assert.NoError(t, err)
	assert.Equal(t, "CREATE INDEX [IDX_users__name] ON [users] ([name])", script)
}
func TestMigratorToDb(t *testing.T) {
	sqlServerDns := "sqlserver://sa:123456@localhost?database=a001&fetchSize=10000&encrypt=disable"
	db, err := tenantDB.Open("mssql", sqlServerDns)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	migrator, err := migrate.NewMigrator(db)
	assert.NoError(t, err)
	scripts, err := migrator.GetSqlMigrate(reflect.TypeOf(User3{}))
	assert.NoError(t, err)
	for _, script := range scripts {
		_, err := db.Exec(script)
		assert.NoError(t, err)
	}
}
