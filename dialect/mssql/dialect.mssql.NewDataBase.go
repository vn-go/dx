package mssql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/internal"
)

func (d *mssqlDialect) NewDataBase(db *db.DB, dbName string) (string, error) {
	sql := `IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = '%s')
			BEGIN
				CREATE DATABASE [%s];
			END;`
	sql = fmt.Sprintf(sql, dbName, dbName)
	_, err := db.Exec(sql)
	if err != nil {
		return "", err
	}
	items := strings.Split(internal.GetDsn(db.Dsn), "&")
	if len(items) > 1 {
		items[0] = strings.Split(items[0], "?")[0] // remove any existing query parameters
		dsn := items[0] + "?database=" + dbName + "&" + items[1]
		return dsn, nil
	} else {
		items[0] = strings.Split(items[0], "?")[0] // remove any existing query parameters
		dsn := items[0] + "?database=" + dbName
		return dsn, nil
	}

}
