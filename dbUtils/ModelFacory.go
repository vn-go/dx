package dbutils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

type modelFacoryType struct {
}
type initGetColsOfModelHasDefaultValue struct {
	cols []entity.ColumnDef
	err  error
	once sync.Once
}

var cacheGetColsOfModelHasDefaultValue sync.Map

func (m *modelFacoryType) getColsOfModelHasDefaultValue(typ reflect.Type) ([]entity.ColumnDef, error) {
	actually, _ := cacheGetColsOfModelHasDefaultValue.LoadOrStore(typ, &initGetColsOfModelHasDefaultValue{})
	item := actually.(*initGetColsOfModelHasDefaultValue)
	item.once.Do(func() {
		model, err := model.ModelRegister.GetModelByType(typ)
		if err != nil {
			item.err = err
			return
		}

		cols := []entity.ColumnDef{}
		for _, col := range model.Entity.Cols {
			if col.Default != "" {
				cols = append(cols, col)
			}
		}
		item.cols = cols

	})
	return item.cols, item.err

}

type initResolveFixDefautValue struct {
	val  *reflect.Value
	err  error
	once sync.Once
}

func (m *modelFacoryType) convert(str string, typ reflect.Type) (interface{}, error) {
	switch typ.Kind() {
	case reflect.Int:
		val, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return nil, err
		}
		return int(val), nil
	case reflect.Int8:
		val, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return nil, err
		}
		return int8(val), nil
	case reflect.Int16:
		val, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return nil, err
		}
		return int16(val), nil
	case reflect.Int32:
		val, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return nil, err
		}
		return int32(val), nil
	case reflect.Int64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Uint:
		val, err := strconv.ParseUint(str, 10, 0)
		if err != nil {
			return nil, err
		}
		return uint(val), nil
	case reflect.Uint8:
		val, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return nil, err
		}
		return uint8(val), nil
	case reflect.Uint16:
		val, err := strconv.ParseUint(str, 10, 16)
		if err != nil {
			return nil, err
		}
		return uint16(val), nil
	case reflect.Uint32:
		val, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return nil, err
		}
		return uint32(val), nil
	case reflect.Uint64:
		val, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Float32:
		val, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return nil, err
		}
		return float32(val), nil
	case reflect.Float64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Bool:
		val, err := strconv.ParseBool(str)
		if err != nil {
			return nil, err
		}
		return val, nil

	default:
		return nil, fmt.Errorf("can not convert %s to %s", str, typ.String())
	}
}

var cacheResolveFixDefautValue sync.Map

func (m *modelFacoryType) resolveFixDefautValue(defaultValua string, fieldType reflect.Type) (*reflect.Value, error) {
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	key := defaultValua + "//" + fieldType.String()
	actually, _ := cacheResolveFixDefautValue.LoadOrStore(key, &initResolveFixDefautValue{})
	item := actually.(*initResolveFixDefautValue)
	item.once.Do(func() {

		if fieldType.Kind() == reflect.String {
			//\'(.*)\'
			reg := regexp.MustCompile(`\'(.*)\'`)
			txt := defaultValua
			matches := reg.FindStringSubmatch(defaultValua)
			if len(matches) > 1 {
				txt = matches[1]
			}
			val := reflect.ValueOf(txt)
			item.val = &val
			return
		} else if number, err := m.convert(defaultValua, fieldType); err == nil {
			val := reflect.ValueOf(number)
			item.val = &val
			return
		}
		item.err = errors.NewUnsupportedError("default value not support: " + defaultValua)
	})
	return item.val, item.err
}
func (m *modelFacoryType) resolveDefautValue(modelType reflect.Type, defaultValua string, field reflect.StructField) (*reflect.Value, error) {
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	var val reflect.Value
	if !strings.HasSuffix(defaultValua, "()") {
		return m.resolveFixDefautValue(defaultValua, fieldType)
	}
	switch defaultValua {
	case "now()":
		if fieldType == reflect.TypeOf(time.Time{}) {

			val = reflect.ValueOf(time.Now().UTC())
			return &val, nil
		} else {
			return nil, errors.NewDFErr(field, modelType, defaultValua)
		}

	case "true":
		if fieldType.Kind() == reflect.Bool {
			val = reflect.ValueOf(true)
			return &val, nil
		} else {
			return nil, errors.NewDFErr(field, modelType, defaultValua)
		}

	case "false":
		if fieldType.Kind() == reflect.Bool {
			val = reflect.ValueOf(false)
			return &val, nil
		} else {
			return nil, errors.NewDFErr(field, modelType, defaultValua)
		}
	case "uuid()":
		if fieldType.Kind() == reflect.String {
			val = reflect.ValueOf(uuid.New().String())
			return &val, nil
		}
		if fieldType == reflect.TypeOf(uuid.UUID{}) {
			val = reflect.ValueOf(uuid.New())
			return &val, nil
		} else {

			return nil, errors.NewDFErr(field, modelType, defaultValua)
		}

	default:
		{
			return nil, errors.NewUnsupportedError("default value not support: " + defaultValua)
		}
	}
}
func (m *modelFacoryType) SetDefaultValue(valOfModel reflect.Value) (interface{}, error) {

	typ := valOfModel.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	cols, err := m.getColsOfModelHasDefaultValue(typ)
	if err != nil {
		return nil, err
	}
	for _, col := range cols {

		valueOfDefaultValue, err := m.resolveDefautValue(typ, col.Default, col.Field)
		if err != nil {
			return nil, err
		}
		fieldVal := valOfModel.Elem().FieldByIndex(col.IndexOfField)
		if fieldVal.IsValid() {

			if fieldVal.CanConvert((*valueOfDefaultValue).Type()) {
				fieldVal.Set(*valueOfDefaultValue)
			}

		}

	}
	ret := valOfModel.Interface()
	return ret, nil
}
