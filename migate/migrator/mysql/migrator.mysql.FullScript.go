package mysql

import (
	"fmt"
	"sync"

	"github.com/vn-go/dx/model"
)

type mysqlGetFullScriptInit struct {
	once sync.Once
	ret  []string
}

func (m *MigratorMySql) GetFullScript() ([]string, error) {
	key := fmt.Sprintf("%s_%s", m.db.DbName, m.db.DriverName)
	actual, _ := m.cacheGetFullScript.LoadOrStore(key, &mysqlGetFullScriptInit{})
	init := actual.(*mysqlGetFullScriptInit)
	var err error
	init.once.Do(func() {
		init.ret, err = m.getFullScript()
	})
	return init.ret, err
}
func (m *MigratorMySql) getFullScript() ([]string, error) {
	sqlInstall, err := m.GetSqlInstallDb()
	if err != nil {
		return nil, err
	}
	scripts := sqlInstall
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlCreateTable(entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}

	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlAddColumn(entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlAddIndex(entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlAddUniqueIndex(entity.Entity.EntityType)
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
