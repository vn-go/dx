package mysql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/internal"
)

func (d *mySqlDialect) NewDataBase(db *db.DB, dbName string) (string, error) {
	cmd := "CREATE DATABASE IF NOT EXISTS `%s`;"
	cmd = fmt.Sprintf(cmd, dbName)
	_, err := db.Exec(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to create database %s: %v", dbName, err)
	}
	items := strings.Split(internal.GetDsn(db.Dsn), "?")
	items[0] = strings.Split(items[0], "/")[0]

	dsn := items[0] + "/" + dbName + "?" + items[1]
	return dsn, nil
}
