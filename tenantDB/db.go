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

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/internal"
)

type TenantDB struct {
	*sql.DB
	info *internal.TenantDBInfo
}

type TenantTx struct {
	*sql.Tx
	info *internal.TenantDBInfo
	Db   *TenantDB
}

func (tx *TenantTx) GetDriverName() string {
	return tx.info.DriverName
}
func (tx *TenantTx) GetDBName() string {
	return tx.info.DbName
}
func (tx *TenantTx) GetDbType() internal.DB_DRIVER_TYPE {
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
	return db.info.DriverName
}
func (db *TenantDB) GetDBName() string {
	return db.info.DbName
}
func (db *TenantDB) GetDbType() internal.DB_DRIVER_TYPE {
	return db.info.DbType
}

type isManagerDb func(driver string, dbName string) bool

var IsManagerDb isManagerDb

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
		info := &internal.TenantDBInfo{
			DriverName: realDriverName,
			Key:        key,
		}

		err = info.Detect(DB)
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
	if db.info == nil {
		db.info = &internal.TenantDBInfo{}
	}
	err := db.info.Detect(db.DB)
	if err != nil {
		return err
	}

	return nil
}
func (db *TenantDB) GetDbVersion() string {
	if !db.info.HasDetected {
		err := db.Detect()
		if err != nil {
			return "error: can not detect database version"
		}
	}
	if db.info.DriverName == "postgres" {
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
