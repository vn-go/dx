package mysql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/model"
)

type mysqlGetFullScriptInit struct {
	once sync.Once
	ret  []string
}

func (m *MigratorMySql) GetFullScript(db *db.DB) ([]string, error) {
	key := fmt.Sprintf("%s_%s", db.DbName, db.DriverName)
	actual, _ := m.cacheGetFullScript.LoadOrStore(key, &mysqlGetFullScriptInit{})
	init := actual.(*mysqlGetFullScriptInit)
	var err error
	init.once.Do(func() {
		init.ret, err = m.getFullScript(db)
	})
	return init.ret, err
}
func (m *MigratorMySql) getFullScript(db *db.DB) ([]string, error) {
	sqlInstall, err := m.GetSqlInstallDb()
	if err != nil {
		return nil, err
	}
	scripts := sqlInstall
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlCreateTable(db, entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, strings.Split(script, ";")...)
		}

	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlAddColumn(db, entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, strings.Split(script, ";\n")...)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlAddIndex(db, entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, strings.Split(script, ";\n")...)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		script, err := m.GetSqlAddUniqueIndex(db, entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, strings.Split(script, ";\n")...)
		}
	}
	scriptForeignKey, err := m.GetSqlAddForeignKey(db)
	if err != nil {
		return nil, err
	}
	scripts = append(scripts, scriptForeignKey...)

	return scripts, nil
}
