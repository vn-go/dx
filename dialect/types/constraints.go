package types

import (
	"fmt"
	"strings"
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
		uk := model.Entity.UniqueConstraints
		if _, ok := uk[name]; ok {
			ret := UKConstraintResult{
				TableName: model.Entity.TableName,
				Columns:   uk[name].Cols,
			}
			for _, col := range uk[name].Cols {
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
type initReplaceStar struct {
	once sync.Once
	val  string
}

var replaceStarCache sync.Map

func ReplaceStarWithCache(driver string, raw string, matche byte, replace byte) string {
	key := fmt.Sprintf("%s_%s_%d_%d", driver, raw, matche, replace)
	actual, _ := replaceStarCache.LoadOrStore(key, &initReplaceStar{})
	init := actual.(*initReplaceStar)
	init.once.Do(func() {
		init.val = ReplaceStar(driver, raw, matche, replace)
	})
	return init.val

}
func ReplaceStar(driver string, raw string, matche byte, replace byte) string {
	var builder strings.Builder
	n := len(raw)
	for i := 0; i < n; i++ {
		if raw[i] == matche {
			if i == 0 || raw[i-1] != '\\' {
				builder.WriteByte(replace)
			} else {
				builder.WriteByte(matche)
			}
		} else {
			builder.WriteByte(raw[i])
		}
	}
	return builder.String()
}
