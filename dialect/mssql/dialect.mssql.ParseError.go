package mssql

import (
	"fmt"
	"strings"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migate/loader/types"
	"github.com/vn-go/dx/model"
)

func (d *mssqlDialect) ParseError(dbSchema *types.DbSchema, err error) error {
	//go-mssqldb.Error
	if mssqlErr, ok := err.(mssql.Error); ok {
		switch mssqlErr.Number {
		case 2627:
			return d.Error2627(dbSchema, mssqlErr)
		case 2628:
			return d.Error2628(dbSchema, mssqlErr)
		case 102:
			return &errors.DbErr{
				Err:          err,
				ErrorType:    errors.ERR_SYNTAX,
				ErrorMessage: "Error syntax near",
			}
		default:
			panic(fmt.Sprintf("unhandled error: %v,code=%d,see file %s", mssqlErr, mssqlErr.Number, `dialect\mssql\dialect.mssql.ParseError.go`))
		}

	}

	panic(fmt.Sprintf("unhandled error: %v,see file %s", err, `dialect\mssql\dialect.mssql.ParseError.go`))
}
func (d *mssqlDialect) Error2628(dbSchema *types.DbSchema, err mssql.Error) error {
	// //"String or binary data would be truncated in table 'hrm.dbo.track_filters', column 'ds_name'. Truncated value: 'u'."
	tableName := strings.Split(err.Message, "'")[1]
	tableName = strings.Split(tableName, "'")[0]
	tableName = strings.Split(tableName, ".")[len(strings.Split(tableName, "."))-1]
	columnName := strings.Split(strings.Split(err.Message, "column '")[1], "'")[0]
	ent := model.ModelRegister.FindEntityByName(tableName)
	if ent == nil {
		return &errors.DbErr{

			Err:          err,
			ErrorType:    errors.ERR_LIMIT_FIELD_SIZE,
			ErrorMessage: err.Message,
			DbCols:       []string{columnName},

			Table: tableName,
		}
		//retConstraint.ModelName = entityRet.EntityType.Name()
		//return fmt.Errorf("mssql error: %s, table: %s, column: %s", err.Message, tableName, columnName)
	}

	//return fmt.Errorf("mssql error: %s, table: %s, column: %s", err.Message, tableName, columnName)
	ret := &errors.DbErr{
		StructName:   ent.EntityType.Name(),
		Err:          err,
		ErrorType:    errors.ERR_LIMIT_FIELD_SIZE,
		ErrorMessage: err.Message,
		DbCols:       []string{columnName},
		Fields:       []string{},
		Table:        tableName,
	}
	for _, col := range ent.Cols {
		if col.Name == columnName {
			ret.Fields = append(ret.Fields, col.Name)
		}
	}
	return ret

}
