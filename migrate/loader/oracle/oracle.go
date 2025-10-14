package oracle

import (
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

type MigratorOracle struct {
	cacheLoadFullSchema sync.Map
}

type initPostgresLoadFullSchema struct {
	once sync.Once
	val  *types.DbSchema
	err  error
}

func (m *MigratorOracle) LoadFullSchema(db *db.DB, schema string) (*types.DbSchema, error) {
	cacheKey := db.DbName + "@" + db.DriverName
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initPostgresLoadFullSchema{})
	init := actual.(*initPostgresLoadFullSchema)
	init.once.Do(func() {
		init.val, init.err = m.loadFullSchema(db, schema)
	})
	return init.val, init.err
}
func (m *MigratorOracle) loadFullSchema(db *db.DB, schema string) (*types.DbSchema, error) {
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
	tables, err := m.LoadAllTable(db, schema)
	if err != nil {
		return nil, err
	}
	pks, err := m.LoadAllPrimaryKey(db, schema)
	if err != nil {
		return nil, err
	}
	uks, err := m.LoadAllUniIndex(db, schema)
	if err != nil {
		return nil, err
	}
	idxs, err := m.LoadAllIndex(db, schema)
	if err != nil {
		return nil, err
	}

	dbName := db.DbName
	schemaData := &types.DbSchema{
		DbName:      dbName,
		Tables:      make(map[string]map[string]bool),
		PrimaryKeys: pks,
		UniqueKeys:  uks,
		Indexes:     idxs,
	}
	foreignKeys, err := m.LoadForeignKey(db, schema)
	if err != nil {
		return nil, err
	}
	schemaData.ForeignKeys = map[string]types.DbForeignKeyInfo{}
	for _, fk := range foreignKeys {
		schemaData.ForeignKeys[fk.ConstraintName] = fk
	}
	for table, columns := range tables {
		cols := make(map[string]bool)
		for col := range columns {
			cols[col] = true
		}
		schemaData.Tables[table] = cols
	}

	return schemaData, nil
}
func (m *MigratorOracle) GetDefaultSchema() string {
	return "app"
}

var MigratorOracleInstance = &MigratorOracle{
	cacheLoadFullSchema: sync.Map{},
}

func NewOracleSchemaLoader() types.IMigratorLoader {

	return MigratorOracleInstance
}
