package postgres

import (
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

type MigratorLoaderPostgres struct {
	cacheLoadFullSchema sync.Map
}

type initPostgresLoadFullSchema struct {
	once sync.Once
	val  *types.DbSchema
	err  error
}

func (m *MigratorLoaderPostgres) LoadFullSchema(db *db.DB, schema string) (*types.DbSchema, error) {
	cacheKey := db.DbName + "@" + db.DriverName + "/" + schema
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initPostgresLoadFullSchema{})
	init := actual.(*initPostgresLoadFullSchema)
	init.once.Do(func() {
		init.val, init.err = m.loadFullSchema(db, schema)
	})
	return init.val, init.err
}
func (m *MigratorLoaderPostgres) loadFullSchema(db *db.DB, schema string) (*types.DbSchema, error) {
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

	dbName := db.DbName
	schemaData := &types.DbSchema{
		DbName: dbName,
		Tables: make(map[string]map[string]bool),
		Db:     db,
	}

	schemaData.Refresh = func() error {
		db := schemaData.Db
		tables, err := m.LoadAllTable(db, schema)
		if err != nil {
			return err
		}
		pks, err := m.LoadAllPrimaryKey(db, schema)
		if err != nil {
			return err
		}
		uks, err := m.LoadAllUniIndex(db, schema)
		if err != nil {
			return err
		}
		idxs, err := m.LoadAllIndex(db, schema)
		if err != nil {
			return err
		}

		foreignKeys, err := m.LoadForeignKey(db, schema)
		if err != nil {
			return err
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
		schemaData.PrimaryKeys = pks
		schemaData.UniqueKeys = uks
		schemaData.Indexes = idxs
		return nil
	}
	err := schemaData.Refresh()
	if err != nil {
		return nil, err
	}
	return schemaData, nil
}
func (m *MigratorLoaderPostgres) GetDefaultSchema() string {
	return "public"
}

var migratorLoaderPostgresInstance = &MigratorLoaderPostgres{
	cacheLoadFullSchema: sync.Map{},
}

func NewPosgresSchemaLoader() types.IMigratorLoader {

	return migratorLoaderPostgresInstance
}
