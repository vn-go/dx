package postgres

import (
	"database/sql"
	"strings"
)

func (d *PostgresDialect) NewDataBase(db *sql.DB, sampleDsn string, dbName string) (string, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	err := db.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return "", err
	}
	if !exists {
		_, err := db.Exec(`CREATE DATABASE ` + dbName + ``)
		if err != nil {
			return "", err
		}
	}
	items := strings.Split(sampleDsn, "?")
	if len(items) > 1 {
		subItems := strings.Split(items[0], "/")
		subItems[len(subItems)-1] = dbName // replace the last item with the new dbName
		dsn := strings.Join(subItems, "/") + "?" + items[1]
		return dsn, nil
	} else {
		subItems := strings.Split(items[0], "/")
		subItems[len(subItems)-1] = dbName // replace the last item with the new dbName
		dsn := strings.Join(subItems, "/")
		return dsn, nil
	}

}
