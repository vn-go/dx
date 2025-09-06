package dx

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	dxErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (m *modelType) Select(fields ...string) *modelTypeSelect {

	return &modelTypeSelect{
		modelType: *m,
		fields:    fields,
	}
}

type initBuildUpdateSqlOnce struct {
	val  string
	err  error
	once sync.Once
}

func (m *modelTypeSelect) buildUpdateSql() (string, error) {

	setterItems := []string{}
	ent, err := model.ModelRegister.GetModelByType(m.typEle)

	if err != nil {
		return "", err
	}
	for _, fieldName := range m.fields {
		setterItems = append(setterItems, fmt.Sprintf("%s.%s=?", ent.Entity.TableName, fieldName))

	}

	whereItems := []string{}
	for _, v := range ent.Entity.PrimaryConstraints {
		for _, f := range v {
			whereItems = append(whereItems, fmt.Sprintf("%s.%s=?", ent.Entity.TableName, f.Name))
		}
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
	err = compiler.BuildWhere(strings.Join(whereItems, " AND "))
	if err != nil {
		if err != nil {
			return "", err
		}
	}
	sql += " WHERE " + compiler.Content

	return sql, nil
}
func (m *modelTypeSelect) Update(data any) UpdateResult {

	argsUpdate := []interface{}{}
	valueOfData := reflect.ValueOf(data).Elem()
	typeOfdata := reflect.TypeOf(data)

	ent, err := model.ModelRegister.GetModelByType(m.typEle)

	if err != nil {
		if err != nil {
			return UpdateResult{
				Error: err,
			}
		}
	}
	for i, fieldName := range m.fields {

		if fieldIndex, fieldType, found := internal.Helper.FindField(typeOfdata, fieldName); found {
			if _, modelFieldTYpe, foundInMode := internal.Helper.FindField(m.typEle, fieldName); foundInMode {

				if !modelFieldTYpe.ConvertibleTo(fieldType) {
					return UpdateResult{
						RowsAffected: 0,
						Error:        dxErrors.NewSysError(fmt.Sprintf("%s.%s can not convert to %s.%s", reflect.TypeOf(data).String(), fieldName, m.typ.String(), fieldName)),
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
					Error:        dxErrors.NewSysError(fmt.Sprintf("%s was not found in %s", fieldName, m.typ.String())),
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
	key := m.typEle.String() + strings.Join(m.fields, "-") + "@modelTypeSelect/Update"
	sql, err := internal.OnceCall(key, func() (string, error) {
		return m.buildUpdateSql()
	})
	if err != nil {
		return UpdateResult{
			Error: err,
		}
	}

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
