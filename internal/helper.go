package internal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type helperType struct {
	SkipDefaulValue string
}

// if s is "true" or "false" retun true
func (m *helperType) IsBool(s string) bool {
	return strings.ToLower(s) == "true" || strings.ToLower(s) == "false"
}
func (m *helperType) IsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
func (m *helperType) IsFloatNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func (m *helperType) GetDefaultValue(defaultValue string, defaultValueByFromDbTag map[string]string) (string, error) {
	if strings.Contains(defaultValue, "'") {
		return defaultValue, nil
	}
	if m.IsFloatNumber(defaultValue) {
		return defaultValue, nil

	} else if m.IsNumber(defaultValue) {
		return defaultValue, nil

	} else if m.IsBool(defaultValue) {
		return defaultValue, nil

	} else if val, ok := defaultValueByFromDbTag[defaultValue]; ok {
		return val, nil
	} else {
		return "", fmt.Errorf("not support default value from %s, review GetGetDefaultValueByFromDbTag() function in %s", defaultValue, reflect.TypeOf(m).Elem())
	}
}

var Helper = &helperType{
	SkipDefaulValue: "vdb::skip",
}
