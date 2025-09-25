package postgres

import (
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/internal"
)

func (d *postgresDialect) NewDataBase(db *db.DB, dbName string) (string, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	err := db.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return "", err
	}
	if !exists {
		_, err := db.Exec(`CREATE DATABASE "` + dbName + `"`)
		if err != nil {
			return "", err
		}
	}
	items := strings.Split(internal.GetDsn(db.Dsn), "?")
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
