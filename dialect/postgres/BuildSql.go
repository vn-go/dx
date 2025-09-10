package postgres

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

func (mssql *postgresDialect) BuildSql(info *types.SqlInfo) (*types.SqlParse, error) {
	panic(fmt.Sprintf("not implement in %s", `dialect\mssql\BuildSql.go`))
}
