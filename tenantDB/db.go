package tenantDB

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

type TenantDB struct {
	*sql.DB
	info *tenantDBInfo
}
type tenantDBInfo struct {
	dbName string

	driverName string
	DbType     DB_DRIVER_TYPE

	Version     string
	hasDetected bool
	key         string
}
type TenantTx struct {
	*sql.Tx
	info *tenantDBInfo
	Db   *TenantDB
}

func (tx *TenantTx) GetDriverName() string {
	return tx.info.driverName
}
func (tx *TenantTx) GetDBName() string {
	return tx.info.dbName
}
func (tx *TenantTx) GetDbType() DB_DRIVER_TYPE {
	return tx.info.DbType
}
func (db *TenantDB) Begin() (*TenantTx, error) {
	if err := db.Detect(); err != nil {
		return nil, err
	}
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	return &TenantTx{
		Tx:   tx,
		info: db.info,
		Db:   db,
	}, nil

}
func (db *TenantDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*TenantTx, error) {
	if err := db.Detect(); err != nil {
		return nil, err
	}
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &TenantTx{
		Tx:   tx,
		info: db.info,
	}, nil
}

func (db *TenantDB) GetDriverName() string {
	return db.info.driverName
}
func (db *TenantDB) GetDBName() string {
	return db.info.dbName
}
func (db *TenantDB) GetDbType() DB_DRIVER_TYPE {
	return db.info.DbType
}

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

type DbDetector struct {
	cacheDetectDatabaseType sync.Map
}
type dbDetectInit struct {
	once sync.Once
	val  tenantDBInfo
	err  error
}

var cacheDbDetector sync.Map

func (info *tenantDBInfo) Detect(db *sql.DB) (*tenantDBInfo, error) {

	key := info.driverName + ":" + info.key
	actual, _ := cacheDbDetector.LoadOrStore(key, &dbDetectInit{})
	di := actual.(*dbDetectInit)
	di.once.Do(func() {
		di.val = tenantDBInfo{
			driverName: info.driverName,
			key:        info.key,
		}
		di.err = di.val.detect(db)
	})
	if di.err != nil {
		return nil, di.err
	} else {

		return &di.val, nil
	}
}

type isManagerDb func(driver string, dbName string) bool

var IsManagerDb isManagerDb

func (info *tenantDBInfo) detect(db *sql.DB) error {

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

	if _, ok := sqlGetDbName[info.driverName]; ok {
		sqlGetDbName := sqlGetDbName[info.driverName]
		var dbNameString sql.NullString
		err := db.QueryRow(sqlGetDbName).Scan(&dbNameString)
		if err != nil {

			return err
		}
		if dbNameString.Valid {
			dbName = dbNameString.String
		}
		if IsManagerDb != nil && IsManagerDb(info.driverName, dbName) {
			dbName = ""
		}
	} else {
		return fmt.Errorf("unsupported database type: %s", string(info.driverName))
	}
	if _, ok := sqlGetDbName[info.driverName]; ok {
		err = db.QueryRow(sqlGetVersion[info.driverName]).Scan(&version)
		if err != nil {
			return err
		}

	}
	info.dbName = dbName
	info.Version = version
	info.DbType = dbTypeMap[info.driverName]
	info.hasDetected = true
	return nil
}

var dbSupport map[string]string = map[string]string{
	"sqlserver": "sqlserver",
	"mssql":     "sqlserver",
	"mysql":     "mysql",
	"pg":        "postgres",
	"postgres":  "postgres",
}

func Open(driverName, dsn string) (*TenantDB, error) {
	if realDriverName, ok := dbSupport[strings.ToLower(driverName)]; ok {
		DB, err := sql.Open(driverName, dsn)
		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256([]byte(dsn))
		// Truncate nếu cần, ví dụ lấy 16 byte đầu (32 hex chars)
		key := hex.EncodeToString(hash[:16])
		info := &tenantDBInfo{
			driverName: realDriverName,
			key:        key,
		}

		info, err = info.Detect(DB)
		if err != nil {
			return nil, err
		}
		ret := &TenantDB{
			DB:   DB,
			info: info,
		}
		if err != nil {
			return nil, err
		}

		return ret, nil
	} else {
		dbList := ""
		for k, v := range dbSupport {
			dbList += k + "-->" + v + "\n"
		}
		return nil, fmt.Errorf("unsupport %s, beloe supported list:\n%s", driverName, dbList)
	}
}

func (db *TenantDB) Detect() error {
	info, err := db.info.Detect(db.DB)
	if err != nil {
		return err
	}
	db.info = info
	return nil
}
func (db *TenantDB) GetDbVersion() string {
	if !db.info.hasDetected {
		err := db.Detect()
		if err != nil {
			return "error: can not detect database version"
		}
	}
	if db.info.driverName == "postgres" {
		re := regexp.MustCompile(`PostgreSQL\s+(\d+)`)
		match := re.FindStringSubmatch(db.info.Version)
		if len(match) > 1 {
			db.info.Version = match[1]

		}
	}

	return db.info.Version
}
func (db *TenantDB) GetMajorVersion() (int, error) {
	ret, err := strconv.Atoi(db.GetDbVersion())
	if err != nil {
		return 0, fmt.Errorf("can not convert %s to int", db.GetDbVersion())
	}
	return ret, nil

}
