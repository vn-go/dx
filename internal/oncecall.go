package internal

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	dxErrors "github.com/vn-go/dx/errors"
)

type oneCallItem[TResult any] struct {
	Val TResult
	Err error
}
type initOnceCall[Tkey any, TResult any] struct {
	Once sync.Once
	Item *oneCallItem[TResult]
}

var cacheOnceCall sync.Map

func OnceCall[Tkey any, TResult any](key Tkey, fn func() (TResult, error)) (TResult, error) {
	actual, _ := cacheOnceCall.LoadOrStore(key, &initOnceCall[Tkey, TResult]{})
	onceCall := actual.(*initOnceCall[Tkey, TResult])
	onceCall.Once.Do(func() {
		onceCall.Item = &oneCallItem[TResult]{}
		onceCall.Item.Val, onceCall.Item.Err = fn()
	})
	return onceCall.Item.Val, onceCall.Item.Err
}

var regexpDBSelectFindPlaceHolder = regexp.MustCompile(`\?`)

func ExtractTextsAndArgs(args []any) ([]string, []any, error) {
	if len(args) == 1 {
		if _, ok := args[0].(string); !ok {
			return nil, nil, dxErrors.NewSysError(fmt.Sprintf("%T is not string", args))
		}
		return []string{args[0].(string)}, nil, nil
	}
	strArgs := []string{}
	for _, a := range args {
		if reflect.TypeOf(a) == reflect.TypeFor[string]() {
			strArgs = append(strArgs, a.(string))
		}
	}

	// Tìm tất cả các kết quả khớp pattern
	matches := regexpDBSelectFindPlaceHolder.FindAllStringIndex(strings.Join(strArgs, ","), -1)
	params := make([]interface{}, len(matches))
	if len(matches) > 0 {

		offsetVar := len(args) - len(matches)
		for i := range matches {
			params[i] = args[offsetVar+i]
		}
	}
	selectFields := args[0 : len(args)-len(matches)]
	strFields := []string{}
	for _, x := range selectFields {
		if reflect.TypeOf(x) == reflect.TypeFor[string]() {
			strFields = append(strFields, x.(string))
		} else {
			errMsg := "field placeholder and argument do not correspond"
			errMsg += "\n"
			for _, x := range args {
				errMsg += fmt.Sprintf("%s", x)
			}
			return nil, nil, dxErrors.NewSysError(errMsg)
		}
	}
	return strFields, params, nil
}
