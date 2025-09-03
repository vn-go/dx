package factory

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/common"
	mysqlDialect "github.com/vn-go/dx/dialect/mysql"
)

type dialectFactoryReceiver struct {
	cacheCreate sync.Map
}
type DB_TYPE int

const (
	DB_TYPE_Unknown DB_TYPE = iota
	DB_TYPE_MySQL
	DB_TYPE_Postgres
	DB_TYPE_Greenplum
	DB_TYPE_Cockroach
	DB_TYPE_MariaDB
	DB_TYPE_MSSQL
	DB_TYPE_SQLite
	DB_TYPE_TiDB
	DB_TYPE_Oracle
)

func (u *dialectFactoryReceiver) DetectDatabaseType(db *sql.DB) (DB_TYPE, string, error) {
	var version string
	var comment string

	queries := []struct {
		query string
	}{
		{"SELECT 'version_comment' ,version();"},       // PostgreSQL,  Cockroach, Greenplum
		{"SELECT 'version_comment', @@VERSION;"},       // SQL Server, Sybase
		{"SELECT 'version_comment',sqlite_version();"}, // SQLite
		{"SELECT 'version_comment',tidb_version();"},   // TiDB
		{"SELECT * FROM v$version"},                    // Oracle
		{"SHOW VARIABLES LIKE 'version_comment'"},      //MySQL
	}

	for _, q := range queries {
		err := db.QueryRow(q.query).Scan(&comment, &version)
		if err == nil && version != "" {
			v := strings.ToLower(version)

			switch {
			case strings.Contains(v, "postgres"):
				if strings.Contains(v, "greenplum") {
					return DB_TYPE_Greenplum, version, nil
				}
				return DB_TYPE_Postgres, version, nil
			case strings.Contains(v, "cockroach"):
				return DB_TYPE_Cockroach, version, nil
			case strings.Contains(v, "mysql"):
				if strings.Contains(v, "mariadb") {
					return DB_TYPE_MariaDB, version, nil
				}
				return DB_TYPE_MySQL, version, nil
			case strings.Contains(v, "mariadb"):
				return DB_TYPE_MariaDB, version, nil
			case strings.Contains(v, "microsoft") || strings.Contains(v, "sql server"):
				return DB_TYPE_MSSQL, version, nil
			case strings.Contains(v, "sqlite"):
				return DB_TYPE_SQLite, version, nil
			case strings.Contains(v, "tidb"):
				return DB_TYPE_TiDB, version, nil
			case strings.Contains(v, "oracle"):
				return DB_TYPE_Oracle, version, nil
			}
		}
	}

	return DB_TYPE_Unknown, version, errors.New("unable to detect database type")
}

type dialectCreateInit struct {
	once sync.Once
	val  common.Dialect
}

func (d *dialectFactoryReceiver) create(driverName string) common.Dialect {
	var ret common.Dialect
	switch driverName {
	case "mysql":
		ret = &mysqlDialect.MysqlDialect{}
	case "postgres":

		ret = &postgresDialect{}

	case "mssql":

		ret = &mssqlDialect{}
	case "sqlserver":
		ret = &mssqlDialect{}
	default:
		panic(fmt.Errorf("unsupported driver: %s", driverName))
	}

	return ret
}

func (d *dialectFactoryReceiver) Create(driverName string) common.Dialect {

	actual, _ := d.cacheCreate.LoadOrStore(driverName, &dialectCreateInit{})
	init := actual.(*dialectCreateInit)
	init.once.Do(func() {
		init.val = d.create(driverName)
	})
	return init.val

}

var DialectFactory = &dialectFactoryReceiver{}
