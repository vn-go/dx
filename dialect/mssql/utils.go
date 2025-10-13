package mssql

import (
	"strings"
	"sync"

	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/migate/loader/types"
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

func FindUKConstraint(dbSchema *types.DbSchema, name string) *UKConstraintResult {
	actual, _ := cacheFindUKConstraint.LoadOrStore(name, &initFindUKConstraint{})
	init := actual.(*initFindUKConstraint)
	init.once.Do(func() {
		init.val = findUKConstraint(dbSchema, name)
	})
	return init.val
}

func findUKConstraint(dbSchema *types.DbSchema, name string) *UKConstraintResult {
	var colInfo *types.ColumnsInfo
	if data, ok := dbSchema.UniqueKeys[strings.ToLower(name)]; ok {
		colInfo = &data
	} else if data, ok := dbSchema.PrimaryKeys[strings.ToLower(name)]; ok {
		colInfo = &data
	}
	if colInfo == nil {
		return nil
	}
	dbCols := make([]string, len(colInfo.Columns))
	for i, col := range colInfo.Columns {
		dbCols[i] = col.Name
	}
	retConstraint := UKConstraintResult{
		TableName: colInfo.TableName,
		DbCols:    dbCols,
	}
	entityRet := model.ModelRegister.FindEntityByName(colInfo.TableName)
	if entityRet == nil {

		return &retConstraint
	}
	retConstraint.Columns = []entity.ColumnDef{}
	retConstraint.Fields = []string{}
	for _, col := range entityRet.Cols {
		if col.Name == colInfo.Columns[0].Name {
			retConstraint.Columns = append(retConstraint.Columns, col)
			retConstraint.Fields = append(retConstraint.Fields, col.Field.Name)

		}
	}

	return &retConstraint
}

type FKConstraintResult struct {
	FormTable  string
	ToTable    string
	FromCols   []string
	ToCols     []string
	FromFields []string
	ToFields   []string
}
