package dx

import (
	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
)

type sqlStatementType struct {
	sql  string
	args []any
	db   *DB
}

func (db *DB) Sql(sqlStatement string, args ...any) *sqlStatementType {
	return &sqlStatementType{
		sql:  sqlStatement,
		args: args,
		db:   db,
	}
}
func (sql *sqlStatementType) GetExecSql() (*types.SqlParse, error) {
	key := "sqlStatementType/GetExecSql/" + sql.sql
	return internal.OnceCall(key, func() (*types.SqlParse, error) {
		info, err := compiler.Compile(sql.sql, sql.db.DriverName)
		if err != nil {
			return nil, err
		}

		return factory.DialectFactory.Create(sql.db.DriverName).BuildSql(info)
	})

}
func (sql *sqlStatementType) ScanRow(items any) error {
	if err := internal.Helper.AddrssertSinglePointerToSlice(items); err != nil {
		return err
	}
	sqlExec, err := sql.GetExecSql()
	if err != nil {
		return err
	}
	sql.db.fecthItems(items, sqlExec.Sql, nil, nil, false, sql.args...)
	return nil
}

type globalOptType struct {
	ShowSql bool
}

var Options = globalOptType{
	ShowSql: false,
}
