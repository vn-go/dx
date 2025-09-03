package postgres

import (
	"fmt"
	"sync"

	"github.com/vn-go/dx/migrate/common"
)

type postgresGetFullScriptInit struct {
	once sync.Once
	ret  []string
}

func (m *MigratorPostgres) GetFullScript() ([]string, error) {
	key := fmt.Sprintf("%s_%s", m.Db.GetDBName(), m.Db.GetDbType())
	actual, _ := m.cacheGetFullScript.LoadOrStore(key, &postgresGetFullScriptInit{})
	init := actual.(*postgresGetFullScriptInit)
	var err error
	init.once.Do(func() {
		init.ret, err = m.getFullScript()
	})
	return init.ret, err
}
func (m *MigratorPostgres) getFullScript() ([]string, error) {

	sqlInstall, err := m.GetSqlInstallDb()
	if err != nil {
		return nil, err
	}
	scripts := sqlInstall
	for _, entity := range common.ModelRegistry.GetAllModels() {
		script, err := m.GetSqlCreateTable(entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}

	}
	for _, entity := range common.ModelRegistry.GetAllModels() {

		script, err := m.GetSqlAddColumn(entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range common.ModelRegistry.GetAllModels() {
		//m.GetSqlAddUniqueIndex()
		script, err := m.GetSqlAddUniqueIndex(entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range common.ModelRegistry.GetAllModels() {
		//m.GetSqlAddUniqueIndex()
		script, err := m.GetSqlAddIndex(entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	scriptForeignKey, err := m.GetSqlAddForeignKey()
	if err != nil {
		return nil, err
	}
	scripts = append(scripts, scriptForeignKey...)

	return scripts, nil
}
