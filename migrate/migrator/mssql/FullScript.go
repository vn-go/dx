package mssql

import (
	"fmt"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/model"
)

type mssqlGetFullScriptInit struct {
	once sync.Once
	ret  []string
}

func (m *migratorMssql) GetFullScript(db *db.DB, schema string) ([]string, error) {
	key := fmt.Sprintf("%s_%s,%s", db.DbName, db.DriverName, schema)
	actual, _ := m.cacheGetFullScript.LoadOrStore(key, &mssqlGetFullScriptInit{})
	init := actual.(*mssqlGetFullScriptInit)
	var err error
	init.once.Do(func() {
		init.ret, err = m.getFullScript(db, schema)
	})
	return init.ret, err
}
func (m *migratorMssql) getFullScript(db *db.DB, schema string) ([]string, error) {

	sqlInstall, err := m.GetSqlInstallDb(schema)
	if err != nil {
		return nil, err
	}
	scripts := sqlInstall
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlCreateTable(db, entity.Entity.EntityType, schema)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}

	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlAddColumn(db, entity.Entity.EntityType, schema)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		//m.GetSqlAddUniqueIndex()
		script, err := m.GetSqlAddUniqueIndex(db, entity.Entity.EntityType, schema)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		//m.GetSqlAddUniqueIndex()
		script, err := m.GetSqlAddIndex(db, entity.Entity.EntityType, schema)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	scriptForeignKey, err := m.GetSqlAddForeignKey(db, schema)
	if err != nil {
		return nil, err
	}
	scripts = append(scripts, scriptForeignKey...)

	return scripts, nil
}
