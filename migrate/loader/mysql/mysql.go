package mysql

import (
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
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

func (m *MigratorLoaderMysql) LoadFullSchema(db *db.DB, schema string) (*types.DbSchema, error) {
	cacheKey := db.Info.DbName + "@" + db.DriverName + "$" + schema
	actual, _ := m.cacheLoadFullSchema.LoadOrStore(cacheKey, &initMySqlLoadFullSchema{})
	initSchema := actual.(*initMySqlLoadFullSchema)
	initSchema.once.Do(func() {
		initSchema.schema, initSchema.err = m.loadFullSchema(db, schema)
	})
	return initSchema.schema, initSchema.err
}
func (m *MigratorLoaderMysql) loadFullSchema(db *db.DB, schema string) (*types.DbSchema, error) {
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
		// PrimaryKeys: pks,
		// UniqueKeys:  uks,
		// Indexes:     idxs,
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
			schemaData.ForeignKeys[strings.ToLower(fk.ConstraintName)] = fk
		}
		for table, columns := range tables {
			cols := make(map[string]bool)
			for col := range columns {
				cols[strings.ToLower(col)] = true //mssql ignore case sensitive column name
			}
			schemaData.Tables[strings.ToLower(table)] = cols //mssql ignore case sensitive table name
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
func (m *MigratorLoaderMysql) GetDefaultSchema() string {
	return "public"

}

var migratorLoaderMysqlInstance = &MigratorLoaderMysql{
	cacheLoadFullSchema: sync.Map{},
}

func NewMysqlMigratorLoader() types.IMigratorLoader {

	return migratorLoaderMysqlInstance
}
