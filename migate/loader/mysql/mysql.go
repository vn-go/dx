package migrate

import (
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migate/loader/types"
)

type MigratorLoaderMysql struct {
	cacheLoadFullSchema sync.Map
	db                  *db.DB
}

func (m *MigratorLoaderMysql) GetDbName() string {
	return m.db.Info.DbName
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

func (m *MigratorLoaderMysql) LoadFullSchema() (*types.DbSchema, error) {
	cacheKey := m.db.Info.DbName
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initMySqlLoadFullSchema{})
	initSchema := actual.(*initMySqlLoadFullSchema)
	initSchema.once.Do(func() {
		initSchema.schema, initSchema.err = m.loadFullSchema()
	})
	return initSchema.schema, initSchema.err
}
func (m *MigratorLoaderMysql) loadFullSchema() (*types.DbSchema, error) {

	tables, err := m.LoadAllTable()
	if err != nil {
		return nil, err
	}
	pks, _ := m.LoadAllPrimaryKey()
	uks, _ := m.LoadAllUniIndex()
	idxs, _ := m.LoadAllIndex()

	dbName := m.GetDbName()
	schema := &types.DbSchema{
		DbName:      dbName,
		Tables:      make(map[string]map[string]bool),
		PrimaryKeys: pks,
		UniqueKeys:  uks,
		Indexes:     idxs,
	}
	foreignKeys, err := m.LoadForeignKey()
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

func NewMysqlSchemaLoader(db *db.DB) types.IMigratorLoader {
	return &MigratorLoaderMysql{
		cacheLoadFullSchema: sync.Map{},
		db:                  db,
	}
}
