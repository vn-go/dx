package internal

import (
	"strings"
	"sync"
	"unicode"

	pluralizeLib "github.com/gertd/go-pluralize"
)

var pluralize = pluralizeLib.NewClient()

type utilsTypes struct {
	cachePlural      sync.Map //<-- cache for pluralize
	cacheToSnakeCase sync.Map //<-- cache for SnakeCase
}
type initPluralize struct {
	val  string
	once sync.Once
}

func (u *utilsTypes) Pluralize(txt string) string {
	actually, _ := u.cachePlural.LoadOrStore(txt, &initPluralize{})
	item := actually.(*initPluralize)
	item.once.Do(func() {
		item.val = strings.ToLower(txt)

	})
	return item.val
}

type initSnakeCase struct {
	val  string
	once sync.Once
}

func (u *utilsTypes) SnakeCase(str string) string {
	actually, _ := u.cacheToSnakeCase.LoadOrStore(str, &initSnakeCase{})
	item := actually.(*initSnakeCase)
	item.once.Do(func() {
		var result []rune
		for i, r := range str {
			if i > 0 && unicode.IsUpper(r) &&
				(unicode.IsLower(rune(str[i-1])) || (i+1 < len(str) && unicode.IsLower(rune(str[i+1])))) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		}
		ret := string(result)
		item.val = ret
	})

	return item.val
}

var Utils = &utilsTypes{}
