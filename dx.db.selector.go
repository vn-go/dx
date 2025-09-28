package dx

import (
	"context"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (dbCtx *dbContext) Model(ent any) *modelType {
	typ := reflect.TypeOf(ent)
	typeEle := typ
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Slice {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	return &modelType{
		db:         dbCtx.DB,
		typ:        typ,
		typEle:     typeEle,
		valuaOfEnt: reflect.ValueOf(ent).Elem(),
		ctx:        dbCtx.ctx,
	}
}
func (db *DB) Model(ent any) *modelType {
	typ := reflect.TypeOf(ent)
	typeEle := typ
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Slice {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	ve := reflect.ValueOf(ent)
	if ve.Kind() == reflect.Ptr {
		ve = ve.Elem()
	}
	return &modelType{
		db:         db,
		typ:        typ,
		typEle:     typeEle,
		valuaOfEnt: ve,
	}
}
func (tx *Tx) Model(ent any) *modelType {
	typ := reflect.TypeOf(ent)
	typeEle := typ
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Slice {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	return &modelType{
		db:         tx.db,
		typ:        typ,
		typEle:     typeEle,
		valuaOfEnt: reflect.ValueOf(ent).Elem(),
		tx:         tx,
	}
}
func (m *modelType) WithContext(ctx context.Context) *modelType {
	m.ctx = ctx
	return m
}

// type modelTypeGetSQLKey struct {
// 	typ reflect.Type

// }

func (m *modelType) GetSQL() (string, []any, error) {
	key := m.typEle

	selectSql, err := internal.OnceCall(key, func() (*types.SqlParse, error) {
		var err error
		ent, err := model.ModelRegister.GetModelByType(m.typEle)
		if err != nil {
			return nil, err
		}
		fields := []string{}
		for _, f := range ent.Entity.Cols {
			fields = append(fields, f.Name+" "+f.Field.Name)
		}
		sqlInfo := &types.SqlInfo{
			StrSelect: strings.Join(fields, ","),
			From:      ent.Entity.TableName,
		}

		sql, err := compiler.GetSql(sqlInfo, m.db.DriverName)
		if err != nil {
			return nil, err
		}

		return sql, nil

	})

	return selectSql.Sql, nil, err
}

func (m *modelType) Find() (any, error) {
	sql, _, err := m.GetSQL()
	if err != nil {
		return nil, err
	}
	// sliceType := reflect.SliceOf(m.typEle)

	// sliceTypeValPtr := reflect.New(sliceType)
	ret, err := m.db.fecthItemsOfType(m.typEle, sql, m.ctx, nil, false)
	if err != nil {
		return nil, err
	}

	return ret.Elem().Interface(), err

}
