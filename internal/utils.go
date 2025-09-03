package internal

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	pluralizeLib "github.com/gertd/go-pluralize"
)

var pluralize = pluralizeLib.NewClient()

type utilsReceiver struct {
	cacheSnakeToPascal sync.Map
	cacheToSnakeCase   sync.Map
	cachePlural        sync.Map
	cacheQueryable     sync.Map
	cacheQuoteText     sync.Map
	EXPR               exprUtils
}
type initPlural struct {
	once sync.Once
	val  string
}

func (u *utilsReceiver) Plural(txt string) string {
	actual, _ := u.cachePlural.LoadOrStore(txt, &initPlural{})
	init := actual.(*initPlural)

	init.once.Do(func() {
		txt = strings.ToLower(txt)
		ret := pluralize.Plural(txt)
		init.val = ret
	})
	return init.val
}

type initSnakeToPascal struct {
	once sync.Once
	val  string
}

func (u *utilsReceiver) snakeToPascal(snake string) string {
	// Handle empty string
	if snake == "" {
		return ""
	}

	// Split the string by underscores
	words := strings.Split(strings.ToLower(snake), "_")
	if len(words) == 0 {
		return ""
	}

	// Initialize result
	result := ""

	// Capitalize the first letter of each word
	for _, word := range words {
		if word != "" {
			runes := []rune(word)
			result += string(unicode.ToUpper(runes[0])) + string(runes[1:])
		}
	}

	return result
}
func (u *utilsReceiver) SnakeToPascal(snake string) string {
	actual, _ := u.cacheSnakeToPascal.LoadOrStore(snake, &initSnakeToPascal{})
	init := actual.(*initSnakeToPascal)
	init.once.Do(func() {
		init.val = u.snakeToPascal(snake)
	})
	return init.val
}

type initToSnakeCase struct {
	once sync.Once
	val  string
}

func (u *utilsReceiver) ToSnakeCase(str string) string {
	actual, _ := u.cacheToSnakeCase.LoadOrStore(str, &initToSnakeCase{})
	init := actual.(*initToSnakeCase)

	init.once.Do(func() {
		init.val = u.toSnakeCase(str)
	})
	return init.val
}
func (u *utilsReceiver) toSnakeCase(str string) string {

	var result []rune
	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) &&
			(unicode.IsLower(rune(str[i-1])) || (i+1 < len(str) && unicode.IsLower(rune(str[i+1])))) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	ret := string(result)

	return ret
}

var Utils = &utilsReceiver{
	cacheToSnakeCase: sync.Map{},
	cachePlural:      sync.Map{},
	cacheQueryable:   sync.Map{},
	cacheQuoteText:   sync.Map{},
	EXPR:             *newExprUtils(),
}

func ReplaceStarWithCache(driver string, raw string, matche byte, replace byte) string {
	key := fmt.Sprintf("%s_%s_%d_%d", driver, raw, matche, replace)
	actual, _ := replaceStarCache.LoadOrStore(key, &initReplaceStar{})
	init := actual.(*initReplaceStar)
	init.once.Do(func() {
		init.val = replaceStar(raw, matche, replace)
	})
	return init.val

}

type initReplaceStar struct {
	once sync.Once
	val  string
}

var replaceStarCache sync.Map

func replaceStar(raw string, matche byte, replace byte) string {
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
