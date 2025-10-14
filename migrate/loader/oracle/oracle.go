package oracle

import (
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

type MigratorOraclePostgres struct {
	cacheLoadFullSchema sync.Map
}

type initPostgresLoadFullSchema struct {
	once sync.Once
	val  *types.DbSchema
	err  error
}

func (m *MigratorOraclePostgres) LoadFullSchema(db *db.DB) (*types.DbSchema, error) {
	cacheKey := db.DbName + "@" + db.DriverName
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initPostgresLoadFullSchema{})
	init := actual.(*initPostgresLoadFullSchema)
	init.once.Do(func() {
		init.val, init.err = m.loadFullSchema(db)
	})
	return init.val, init.err
}
func (m *MigratorOraclePostgres) loadFullSchema(db *db.DB) (*types.DbSchema, error) {
	if types.SkipLoadSchemaOnMigrate {
		return &types.DbSchema{
			DbName:      db.DbName,
			Tables:      map[string]map[string]bool{},
			PrimaryKeys: map[string]types.ColumnsInfo{},
			UniqueKeys:  map[string]types.ColumnsInfo{},
			Indexes:     map[string]types.ColumnsInfo{},
			ForeignKeys: map[string]types.DbForeignKeyInfo{},
		}, nil
	}
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

var MigratorOraclePostgresInstance = &MigratorOraclePostgres{
	cacheLoadFullSchema: sync.Map{},
}

func NewOracleSchemaLoader() types.IMigratorLoader {

	return MigratorOraclePostgresInstance
}
