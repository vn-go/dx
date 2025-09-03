package internal

import (
	"database/sql"
	"fmt"
	"sync"
)

type DB_DRIVER_TYPE string

const (
	DB_DRIVER_Postgres  DB_DRIVER_TYPE = "postgres"
	DB_DRIVER_MySQL     DB_DRIVER_TYPE = "mysql"
	DB_DRIVER_MariaDB   DB_DRIVER_TYPE = "mariadb"
	DB_DRIVER_MSSQL     DB_DRIVER_TYPE = "sqlserver"
	DB_DRIVER_SQLite    DB_DRIVER_TYPE = "sqlite"
	DB_DRIVER_Oracle    DB_DRIVER_TYPE = "oracle"
	DB_DRIVER_TiDB      DB_DRIVER_TYPE = "tidb"
	DB_DRIVER_Cockroach DB_DRIVER_TYPE = "cockroach"
	DB_DRIVER_Greenplum DB_DRIVER_TYPE = "greenplum"
	DB_DRIVER_Unknown   DB_DRIVER_TYPE = "unknown"
)

type TenantDBInfo struct {
	DbName string

	DriverName string
	DbType     DB_DRIVER_TYPE

	Version     string
	HasDetected bool
	Key         string
}

func (info *TenantDBInfo) Detect(db *sql.DB) error {

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
		if TenantDbManagerInstance.isManagerDb(info.DriverName, dbName) {
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
	info.HasDetected = true
	return nil
}

type DbDetector struct {
	cacheDetectDatabaseType sync.Map
}
type TenantDbManager struct {
	mapManagerDb map[string]bool
}

var TenantDbManagerInstance = &TenantDbManager{mapManagerDb: make(map[string]bool)}

func (t *TenantDbManager) SetManagerDb(driver string, dbName string) error {
	if driver == "" || dbName == "" {
		return fmt.Errorf("driver and dbName cannot be empty")
	}
	if driver == "postgres" || driver == "mysql" || driver == "sqlite3" || driver == "mssql" {
		t.mapManagerDb[dbName+"://"+driver] = true
	} else {
		return fmt.Errorf("driver %s is not supported for manager db", driver)

	}
	return nil
}
func (t *TenantDbManager) isManagerDb(driver string, dbName string) bool {
	if _, ok := t.mapManagerDb[dbName+"://"+driver]; ok {
		return true
	}
	return false
}
