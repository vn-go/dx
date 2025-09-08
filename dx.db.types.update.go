package dx

import (
	"fmt"
	"reflect"
	"strings"

	dxErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (m *modelType) Select(args ...any) *selectorTypes {

	ret := m.db.Select(args...)
	ret.entityType = &m.typEle
	ret.valuaOfEnt = m.valuaOfEnt
	ret.ctx = m.ctx
	ret.sqlTx = m.tx.Tx
	return ret
}

func (m *selectorTypes) buildUpdateSql() (string, []any, error) {
	strWhere, args := m.getFilter()
	key := (*m.entityType).String() + "://selectorTypes/buildUpdateSql/" + strWhere

	retSql, err := internal.OnceCall(key, func() (string, error) {
		setterItems := []string{}
		ent, err := model.ModelRegister.GetModelByType(*m.entityType)

		if err != nil {
			return "", err
		}
		for _, fieldName := range m.selectFields {
			setterItems = append(setterItems, fmt.Sprintf("%s.%s=?", ent.Entity.TableName, fieldName))

		}

		if strWhere == "" {
			whereItems := []string{}
			for _, v := range ent.Entity.PrimaryConstraints {
				for _, f := range v {
					whereItems = append(whereItems, fmt.Sprintf("%s.%s=?", ent.Entity.TableName, f.Name))
				}
			}
			strWhere = strings.Join(whereItems, " AND ")
		}

		compiler, err := expr.NewExprCompiler(m.db.DB)
		if err != nil {
			return "", err
		}

		compiler.Context.Purpose = expr.BUILD_UPDATE

		err = compiler.BuildSetter(strings.Join(setterItems, ","))
		if err != nil {
			return "", err
		}

		sql := "update " + compiler.Context.Dialect.Quote(ent.Entity.TableName) + " set " + compiler.Content
		compiler.Context.Purpose = expr.BUILD_UPDATE
		err = compiler.BuildWhere(strWhere)
		if err != nil {
			if err != nil {
				return "", err
			}
		}
		sql += " WHERE " + compiler.Content

		return sql, nil
	})
	if err != nil {
		return "", nil, err
	}
	return retSql, args, nil
}
func (m *selectorTypes) Update(data any) UpdateResult {

	argsUpdate := []interface{}{}
	valueOfData := reflect.ValueOf(data).Elem()
	typeOfdata := reflect.TypeOf(data)

	ent, err := model.ModelRegister.GetModelByType(*m.entityType)

	if err != nil {
		if err != nil {
			return UpdateResult{
				Error: err,
			}
		}
	}
	for i, fieldName := range m.selectFields {

		if fieldIndex, fieldType, found := internal.Helper.FindField(typeOfdata, fieldName); found {
			if _, modelFieldTYpe, foundInMode := internal.Helper.FindField(*m.entityType, fieldName); foundInMode {

				if !modelFieldTYpe.ConvertibleTo(fieldType) {
					return UpdateResult{
						RowsAffected: 0,
						Error:        dxErrors.NewSysError(fmt.Sprintf("%s.%s can not convert to %s.%s", reflect.TypeOf(data).String(), fieldName, (*m.entityType).String(), fieldName)),
					}
				} else {
					val := valueOfData.FieldByIndex(fieldIndex)
					if val.Kind() == reflect.Ptr {
						val = val.Elem()
					}
					argsUpdate = append(argsUpdate, val.Interface())
				}
			} else {
				return UpdateResult{
					RowsAffected: 0,
					Error:        dxErrors.NewSysError(fmt.Sprintf("%s was not found in %s", fieldName, (*m.entityType).String())),
				}
			}
			argsUpdate[i] = valueOfData.FieldByIndex(fieldIndex).Interface()
		} else {
			return UpdateResult{
				RowsAffected: 0,
				Error:        dxErrors.NewSysError(fmt.Sprintf("%s was not foudn in %s", fieldName, reflect.TypeOf(data).String())),
			}
		}
	}

	for _, v := range ent.Entity.PrimaryConstraints {
		for _, f := range v {

			val := m.valuaOfEnt.FieldByIndex(f.IndexOfField)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			argsVal := val.Interface()
			argsUpdate = append(argsUpdate, argsVal)

		}

	}
	sql, args, err := m.buildUpdateSql()
	if err != nil {
		return UpdateResult{
			Error: err,
		}
	}
	argsUpdate = append(argsUpdate, args...)
	r, err := m.db.Exec(sql, argsUpdate...)
	if err != nil {
		return UpdateResult{
			Error: err,
			Sql:   sql,
		}
	}
	rn, err := r.RowsAffected()
	if err != nil {
		return UpdateResult{
			Error: err,
			Sql:   sql,
		}
	}
	return UpdateResult{
		RowsAffected: rn,
	}
}
