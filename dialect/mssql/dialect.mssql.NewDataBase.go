package mssql

import (
	"database/sql"
	"fmt"
	"strings"
)

func (d *mssqlDialect) NewDataBase(db *sql.DB, sampleDsn string, dbName string) (string, error) {
	sql := `IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = '%s')
			BEGIN
				CREATE DATABASE [%s];
			END;`
	sql = fmt.Sprintf(sql, dbName, dbName)
	_, err := db.Exec(sql)
	if err != nil {
		return "", err
	}
	items := strings.Split(sampleDsn, "&")
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
