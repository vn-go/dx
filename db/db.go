package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/internal"
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

type Info struct {
	DbName string

	DriverName string
	//DbType     DB_DRIVER_TYPE

	Version string
	Dsn     string
}
type DB struct {
	*sql.DB
	*Info
}
type initEncryptDsn struct {
	val  string
	once sync.Once
}

var cachEncryptDsn sync.Map

func encryptDsn(dsn string) string {
	actually, _ := cachEncryptDsn.LoadOrStore(dsn, &initEncryptDsn{})
	item := actually.(*initEncryptDsn)
	item.once.Do(func() {
		var prefix string = ""
		if strings.Contains(dsn, "://") {
			prefix = strings.Split(dsn, "://")[0]
			dsn = strings.Split(dsn, "://")[1]
		}
		dbPass := strings.Split(strings.Split(dsn, "@")[0], ":")[1]
		item.val = strings.Replace(dsn, ":"+dbPass+"@", ":****@", 1)
		if prefix != "" {
			item.val = prefix + "://" + item.val
		}
	})

	return item.val
}

func extractDsn(driverName, dsn string) string {
	if driverName == "mysql" { // fix mysql dsn do not have multiStatements=true
		if !strings.Contains(dsn, "?") {
			dsn += "?multiStatements=true&parseTime=true"
		} else {
			if !strings.Contains(dsn, "multiStatements=true") {
				dsn += "&multiStatements=true"
			}
			if !strings.Contains(dsn, "parseTime=true") {
				dsn += "&parseTime=true"
			}
		}
	}

	return dsn

}

func Open(driverName, dsn string) (*DB, error) {
	dsn = extractDsn(driverName, dsn)
	dsnEncypt := encryptDsn(dsn)
	ret := &DB{
		Info: &Info{
			DriverName: driverName,
			Dsn:        dsnEncypt,
		},
	}
	DB, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	ret.DB = DB
	err = detect(ret.Info, ret.DB)
	if err != nil {
		return nil, err
	}
	internal.SetDsn(dsnEncypt, dsn)

	return ret, nil
}

func (db *DB) GetDbVersion() string {

	if db.Info.DriverName == "postgres" {
		re := regexp.MustCompile(`PostgreSQL\s+(\d+)`)
		match := re.FindStringSubmatch(db.Info.Version)
		if len(match) > 1 {
			db.Info.Version = match[1]

		}
	}

	return db.Info.Version
}
func (db *DB) GetMajorVersion() (int, error) {
	ret, err := strconv.Atoi(db.GetDbVersion())
	if err != nil {
		return 0, fmt.Errorf("can not convert %s to int", db.GetDbVersion())
	}
	return ret, nil

}
