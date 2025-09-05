package postgres

import (
	"fmt"
	"strings"
	"sync"
)

type initReplaceStar struct {
	once sync.Once
	val  string
}

var replaceStarCache sync.Map

func replaceStarWithCache(driver string, raw string, matche byte, replace byte) string {
	key := fmt.Sprintf("%s_%s_%d_%d", driver, raw, matche, replace)
	actual, _ := replaceStarCache.LoadOrStore(key, &initReplaceStar{})
	init := actual.(*initReplaceStar)
	init.once.Do(func() {
		init.val = replaceStar(driver, raw, matche, replace)
	})
	return init.val

}
func replaceStar(driver string, raw string, matche byte, replace byte) string {
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
