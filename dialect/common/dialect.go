package common

import (
	"database/sql"

	"github.com/vn-go/dx/migrate/common"
	//"github.com/vn-go/xdb/migrate"
	//"github.com/vn-go/xdb/migrate"
)

// Certain functions are transformed during compilation based on the SQL dialect.
// The dialect decides whether a function needs to be adapted for the target database driver.
// If required, the function will be rewritten here.
type DialectDelegateFunction struct {
	FuncName         string
	Args             []string
	HandledByDialect bool // ✅ Indicates if this function is allowed to be delegated to the dialect
}

type Dialect interface {
	LikeValue(val string) string

	ParseError(dbSchame *common.DbSchema, err error) error
	Name() string
	Quote(str ...string) string
	GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error)
	ToText(value string) string
	ToParam(index int) string
	SqlFunction(delegator *DialectDelegateFunction) (string, error)
	MakeSqlInsert(tableName string, columns []common.ColumnDef, data interface{}) (string, []interface{})
	NewDataBase(db *sql.DB, sampleDsn string, dbName string) (string, error)
	MakeSelectTop(sql string, top int) string
}
