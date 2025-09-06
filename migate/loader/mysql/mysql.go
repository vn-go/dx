package mysql

import (
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migate/loader/types"
)

type MigratorLoaderMysql struct {
	cacheLoadFullSchema sync.Map
}

/*
This structure is used to ensure that each database runs only once,
regardless of the Go Routines
*/
type initMySqlLoadFullSchema struct {
	once   sync.Once
	err    error
	schema *types.DbSchema
}

func (m *MigratorLoaderMysql) LoadFullSchema(db *db.DB) (*types.DbSchema, error) {
	cacheKey := db.Info.DbName + "@" + db.DriverName
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initMySqlLoadFullSchema{})
	initSchema := actual.(*initMySqlLoadFullSchema)
	initSchema.once.Do(func() {
		initSchema.schema, initSchema.err = m.loadFullSchema(db)
	})
	return initSchema.schema, initSchema.err
}
func (m *MigratorLoaderMysql) loadFullSchema(db *db.DB) (*types.DbSchema, error) {

	tables, err := m.LoadAllTable(db)
	if err != nil {
		return nil, err
	}
	pks, _ := m.LoadAllPrimaryKey(db)
	uks, _ := m.LoadAllUniIndex(db)
	idxs, _ := m.LoadAllIndex(db)

	dbName := db.DbName
	schema := &types.DbSchema{
		DbName:      dbName,
		Tables:      make(map[string]map[string]bool),
		PrimaryKeys: pks,
		UniqueKeys:  uks,
		Indexes:     idxs,
	}
	foreignKeys, err := m.LoadForeignKey(db)
	if err != nil {
		return nil, err
	}
	schema.ForeignKeys = map[string]types.DbForeignKeyInfo{}
	for _, fk := range foreignKeys {
		schema.ForeignKeys[fk.ConstraintName] = fk
	}
	for table, columns := range tables {
		cols := make(map[string]bool)
		for col := range columns {
			cols[col] = true
		}
		schema.Tables[table] = cols
	}

	return schema, nil
}

var migratorLoaderMysqlInstance = &MigratorLoaderMysql{
	cacheLoadFullSchema: sync.Map{},
}

func NewMysqlMigratorLoader() types.IMigratorLoader {

	return migratorLoaderMysqlInstance
}
