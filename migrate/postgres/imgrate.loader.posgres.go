package postgres

import (
	"sync"

	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

type MigratorLoaderPostgres struct {
	cacheLoadFullSchema sync.Map
}

func (m *MigratorLoaderPostgres) GetDbName(db *tenantDB.TenantDB) string {
	return db.GetDBName()
}

type initPostgresLoadFullSchema struct {
	once sync.Once
	val  *common.DbSchema
	err  error
}

func (m *MigratorLoaderPostgres) LoadFullSchema(db *tenantDB.TenantDB) (*common.DbSchema, error) {
	cacheKey := db.GetDBName()
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initPostgresLoadFullSchema{})
	init := actual.(*initPostgresLoadFullSchema)
	init.once.Do(func() {
		init.val, init.err = m.loadFullSchema(db)
	})
	return init.val, init.err
}
func (m *MigratorLoaderPostgres) loadFullSchema(db *tenantDB.TenantDB) (*common.DbSchema, error) {

	tables, err := m.LoadAllTable(db)
	if err != nil {
		return nil, err
	}
	pks, err := m.LoadAllPrimaryKey(db)
	if err != nil {
		return nil, err
	}
	uks, err := m.LoadAllUniIndex(db)
	if err != nil {
		return nil, err
	}
	idxs, err := m.LoadAllIndex(db)
	if err != nil {
		return nil, err
	}

	dbName := m.GetDbName(db)
	schema := &common.DbSchema{
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
	schema.ForeignKeys = map[string]common.DbForeignKeyInfo{}
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
