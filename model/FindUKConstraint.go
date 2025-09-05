package model

import "github.com/vn-go/dx/entity"

type UKConstraintResult struct {
	TableName string
	Columns   []entity.ColumnDef
	DbCols    []string
	Fields    []string
	//Columns   []string
}

func (r *modelRegister) FindUKConstraint(name string) *UKConstraintResult {
	for _, model := range r.GetAllModels() {
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
