package oracle

import (
	"fmt"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/model"
)

type postgresGetFullScriptInit struct {
	once sync.Once
	ret  []string
}

func (m *MigratorOracle) GetFullScript(db *db.DB) ([]string, error) {
	key := fmt.Sprintf("%s_%s", db.DbName, db.DriverName)
	actual, _ := m.cacheGetFullScript.LoadOrStore(key, &postgresGetFullScriptInit{})
	init := actual.(*postgresGetFullScriptInit)
	var err error
	init.once.Do(func() {
		init.ret, err = m.getFullScript(db)
	})
	return init.ret, err
}
func (m *MigratorOracle) getFullScript(db *db.DB) ([]string, error) {

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
			scripts = append(scripts, script)
		}

	}
	for _, entity := range model.ModelRegister.GetAllModels() {

		script, err := m.GetSqlAddColumn(db, entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		//m.GetSqlAddUniqueIndex()
		script, err := m.GetSqlAddUniqueIndex(db, entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	for _, entity := range model.ModelRegister.GetAllModels() {
		//m.GetSqlAddUniqueIndex()
		script, err := m.GetSqlAddIndex(db, entity.Entity.EntityType)
		if err != nil {
			return nil, err
		}
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	scriptForeignKey, err := m.GetSqlAddForeignKey(db)
	if err != nil {
		return nil, err
	}
	scripts = append(scripts, scriptForeignKey...)

	return scripts, nil
}
