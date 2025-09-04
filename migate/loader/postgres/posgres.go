package postgres

import (
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migate/loader/types"
)

type MigratorLoaderPostgres struct {
	cacheLoadFullSchema sync.Map
	db                  *db.DB
}

func (m *MigratorLoaderPostgres) GetDbName() string {
	return m.db.Info.DbName
}

type initPostgresLoadFullSchema struct {
	once sync.Once
	val  *types.DbSchema
	err  error
}

func (m *MigratorLoaderPostgres) LoadFullSchema() (*types.DbSchema, error) {
	cacheKey := m.GetDbName()
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initPostgresLoadFullSchema{})
	init := actual.(*initPostgresLoadFullSchema)
	init.once.Do(func() {
		init.val, init.err = m.loadFullSchema()
	})
	return init.val, init.err
}
func (m *MigratorLoaderPostgres) loadFullSchema() (*types.DbSchema, error) {

	tables, err := m.LoadAllTable()
	if err != nil {
		return nil, err
	}
	pks, err := m.LoadAllPrimaryKey()
	if err != nil {
		return nil, err
	}
	uks, err := m.LoadAllUniIndex()
	if err != nil {
		return nil, err
	}
	idxs, err := m.LoadAllIndex()
	if err != nil {
		return nil, err
	}

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
func NewPosgresSchemaLoader(db *db.DB) types.IMigratorLoader {

	return &MigratorLoaderPostgres{
		cacheLoadFullSchema: sync.Map{},
		db:                  db,
	}
}
