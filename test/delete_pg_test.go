package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/test/models"
)

func ReplacePlaceholders(dialect, query string) string {
	var builder strings.Builder
	inSingle := false
	inDouble := false
	argIndex := 1

	for i := 0; i < len(query); i++ {
		ch := query[i]

		switch ch {
		case '\'':
			// Toggle trạng thái nếu không bị escape
			if !inDouble {
				inSingle = !inSingle
			}
			builder.WriteByte(ch)

		case '"':
			// Toggle trạng thái nếu không bị escape
			if !inSingle {
				inDouble = !inDouble
			}
			builder.WriteByte(ch)

		case '?':
			if inSingle || inDouble {
				// '?' nằm trong literal, giữ nguyên
				builder.WriteByte('?')
			} else {
				switch dialect {
				case "postgres", "postgresql":
					builder.WriteString(fmt.Sprintf("$%d", argIndex))
				case "sqlserver":
					builder.WriteString(fmt.Sprintf("@p%d", argIndex))
				default:
					builder.WriteByte('?')
				}
				argIndex++
			}

		default:
			builder.WriteByte(ch)
		}
	}

	return builder.String()
}
func TestReplacePlaceholders(t *testing.T) {
	a, b := internal.Helper.InspectStringParam(`select concat(name,'a''?') FName,?+?*?-? where code=? and name='"?"' and desc='c?a?'`)
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(ReplacePlaceholders("postgres", a))
	fmt.Println(ReplacePlaceholders("postgres",
		`select concat(name,'a''?') FName,?+?*?-? where code=? and name='"?"' and desc='c?a?'`))
}
func TestDeleteUserPg(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	err = db.Delete(&models.User{}, "username=?", "admin").Error
	assert.NoError(t, err)
}
