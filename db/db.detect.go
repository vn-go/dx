package db

import (
	"database/sql"
	"fmt"
)

func detect(info *Info, db *sql.DB) error {

	var version string
	dbName := ""

	sqlGetDbName := map[string]string{
		"postgres":  "SELECT current_database()",
		"mysql":     "SELECT DATABASE()",
		"sqlite":    "SELECT name FROM sqlite_master WHERE type='table' AND name='sqlite_sequence'",
		"tidb":      "SELECT DATABASE()",
		"oracle":    "SELECT SYS_CONTEXT('USERENV', 'DB_NAME') FROM dual",
		"cockroach": "SELECT current_database()",
		"greenplum": "SELECT current_database()",
		"sqlserver": "SELECT DB_NAME()",
	}
	sqlGetVersion := map[string]string{
		"postgres":  "SELECT version()",
		"mysql":     "SELECT @@VERSION",
		"sqlite":    "SELECT sqlite_version()",
		"tidb":      "SELECT tidb_version()",
		"oracle":    "SELECT * FROM v$version",
		"cockroach": "SELECT version()",
		"greenplum": "SELECT version()",
		"sqlserver": "SELECT @@VERSION",
	}
	dbTypeMap := map[string]DB_DRIVER_TYPE{
		"postgres":  DB_DRIVER_Postgres,
		"mysql":     DB_DRIVER_MySQL,
		"sqlite":    DB_DRIVER_SQLite,
		"tidb":      DB_DRIVER_TiDB,
		"oracle":    DB_DRIVER_Oracle,
		"cockroach": DB_DRIVER_Cockroach,
		"greenplum": DB_DRIVER_Greenplum,
		"sqlserver": DB_DRIVER_MSSQL,
	}
	err := db.Ping()
	if err != nil {
		return err
	}

	if _, ok := sqlGetDbName[info.DriverName]; ok {
		sqlGetDbName := sqlGetDbName[info.DriverName]
		var dbNameString sql.NullString
		err := db.QueryRow(sqlGetDbName).Scan(&dbNameString)
		if err != nil {

			return err
		}
		if dbNameString.Valid {
			dbName = dbNameString.String
		}
		if IsManagerDb(info.DriverName, dbName) {
			dbName = ""
		}
	} else {
		return fmt.Errorf("unsupported database type: %s", string(info.DriverName))
	}
	if _, ok := sqlGetDbName[info.DriverName]; ok {
		err = db.QueryRow(sqlGetVersion[info.DriverName]).Scan(&version)
		if err != nil {
			return err
		}

	}
	info.DbName = dbName
	info.Version = version
	info.DbType = dbTypeMap[info.DriverName]

	return nil
}

var managerDb = map[string]string{}

func SetManagerDb(driver string, dbName string) {
	managerDb[driver] = dbName

}
func IsManagerDb(driver string, dbName string) bool {
	if dbNameManager, ok := managerDb[driver]; ok && dbName == dbNameManager {
		return true
	}
	return false
}
