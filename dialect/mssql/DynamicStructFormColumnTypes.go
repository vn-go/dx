package mssql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/vn-go/dx/internal"
)

/*
This function will create a dynamic struct based on the column types of a SQL query result.
Note: for sql server, the column types are not always correct, so this function may not work correctly.
*/
func (d *mssqlDialect) DynamicStructFormColumnTypes(sql string, colTypes []*sql.ColumnType) reflect.Type {
	fields := make([]reflect.StructField, 0, len(colTypes))

	for _, col := range colTypes {
		dbType := strings.ToUpper(col.DatabaseTypeName())
		var goType reflect.Type

		switch {
		// -------- INTEGER TYPES --------
		case strings.Contains(dbType, "BIGINT"):
			goType = reflect.TypeFor[*int64]()
		case strings.Contains(dbType, "INT"):
			goType = reflect.TypeFor[*int32]()
		case strings.Contains(dbType, "SMALLINT"), strings.Contains(dbType, "TINYINT"):
			goType = reflect.TypeFor[*int16]()

		// -------- FLOAT / DECIMAL --------
		case strings.Contains(dbType, "FLOAT"), strings.Contains(dbType, "REAL"):
			goType = reflect.TypeFor[*float64]()
		case strings.Contains(dbType, "DECIMAL"), strings.Contains(dbType, "NUMERIC"), strings.Contains(dbType, "MONEY"):
			goType = reflect.TypeFor[*float64]()

		// -------- BOOLEAN --------
		case strings.Contains(dbType, "BIT"):
			goType = reflect.TypeFor[*bool]()

		// -------- STRING TYPES --------
		case strings.Contains(dbType, "CHAR"),
			strings.Contains(dbType, "TEXT"),
			strings.Contains(dbType, "XML"),
			strings.Contains(dbType, "NVARCHAR"),
			strings.Contains(dbType, "NCHAR"),
			strings.Contains(dbType, "VARCHAR"):
			goType = reflect.TypeFor[*string]()

		// -------- DATE / TIME --------
		case strings.Contains(dbType, "DATE"),
			strings.Contains(dbType, "TIME"),
			strings.Contains(dbType, "DATETIME"),
			strings.Contains(dbType, "SMALLDATETIME"):
			goType = reflect.TypeFor[*time.Time]()

		// -------- UNIQUEIDENTIFIER --------
		case strings.Contains(dbType, "UNIQUEIDENTIFIER"):
			goType = reflect.TypeFor[*string]()

		// -------- BINARY TYPES --------
		case strings.Contains(dbType, "BINARY"),
			strings.Contains(dbType, "VARBINARY"),
			strings.Contains(dbType, "IMAGE"):
			goType = reflect.TypeFor[*[]byte]()

		default:
			goType = reflect.TypeFor[*interface{}]()
		}

		field := reflect.StructField{
			Name: strings.Title(col.Name()),
			Type: goType,
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, internal.Helper.ToLowerCamel(col.Name()))),
		}
		fields = append(fields, field)
	}

	return reflect.StructOf(fields)
}
