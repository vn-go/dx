package mssql

import (
	"sync"

	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/model"
)

type initFindUKConstraint struct {
	once sync.Once
	val  *UKConstraintResult
}
type UKConstraintResult struct {
	TableName string
	Columns   []entity.ColumnDef
	DbCols    []string
	Fields    []string
	//Columns   []string
}

var cacheFindUKConstraint sync.Map

func FindUKConstraint(name string) *UKConstraintResult {
	actual, _ := cacheFindUKConstraint.LoadOrStore(name, &initFindUKConstraint{})
	init := actual.(*initFindUKConstraint)
	init.once.Do(func() {
		init.val = findUKConstraint(name)
	})
	return init.val
}

func findUKConstraint(name string) *UKConstraintResult {
	for _, model := range model.ModelRegister.GetAllModels() {
		uk := model.Entity.BuildUniqueConstraints
		if _, ok := uk[name]; ok {
			ret := UKConstraintResult{
				TableName: model.Entity.TableName,
				Columns:   uk[name],
			}
			for _, col := range uk[name] {
				ret.DbCols = append(ret.DbCols, col.Name)
				ret.Fields = append(ret.Fields, col.Field.Name)
			}
			return &ret
		}
	}
	return nil
}

type FKConstraintResult struct {
	FormTable  string
	ToTable    string
	FromCols   []string
	ToCols     []string
	FromFields []string
	ToFields   []string
}
