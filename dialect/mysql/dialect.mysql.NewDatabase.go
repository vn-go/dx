package mysql

import (
	"database/sql"
	"fmt"
	"strings"
)

func (d *mySqlDialect) NewDataBase(db *sql.DB, sampleDsn string, dbName string) (string, error) {
	cmd := "CREATE DATABASE IF NOT EXISTS %s;"
	cmd = fmt.Sprintf(cmd, dbName)
	_, err := db.Exec(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to create database %s: %v", dbName, err)
	}
	items := strings.Split(sampleDsn, "?")
	items[0] = strings.Split(items[0], "/")[0]

	dsn := items[0] + "/" + dbName + "?" + items[1]
	return dsn, nil
}
