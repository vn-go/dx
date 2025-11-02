package dx

import (
	"context"
	"database/sql"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/dialect/types"
)

type DB struct {
	*db.DB
	Dialect types.Dialect
}

type TxOptions struct {
	// Isolation is the transaction isolation level.
	// If zero, the driver or database's default level is used.
	Isolation IsolationLevel
	ReadOnly  bool
}
type IsolationLevel int

// Various isolation levels that drivers may support in [DB.BeginTx].
// If a driver does not support a given isolation level an error may be returned.
//
// See https://en.wikipedia.org/wiki/Isolation_(database_systems)#Isolation_levels.
const (
	LevelDefault IsolationLevel = iota
	LevelReadUncommitted
	LevelReadCommitted
	LevelWriteCommitted
	LevelRepeatableRead
	LevelSnapshot
	LevelSerializable
	LevelLinearizable
)

type Tx struct {
	*sql.Tx
	db    *DB
	Error error
}

func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error) {
	if opts == nil {
		tx, err := db.DB.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &Tx{
			db: db,
			Tx: tx,
		}, nil
	} else {
		tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.IsolationLevel(opts.Isolation),
			ReadOnly:  opts.ReadOnly,
		})
		if err != nil {
			return nil, err
		}
		return &Tx{
			db: db,
			Tx: tx,
		}, nil
	}

}
