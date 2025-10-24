package sql

import (
	"fmt"
	"testing"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func Test0001(t *testing.T) {
	str := `from(
					employee emp,
					user usr,
					department dept,
					emp.department_id like department.id,
					emp.userId = usr.id,
			),emp.code+dept.code Code, where(emp.id = select(max(emp.id)))`
	str = `max(emp.id), from(employee emp)`
	sql, err := smartier.simple(str)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)

}
func BenchmarkTest0001(b *testing.B) {
	str := `from(
					employee emp,
					user usr,
					department dept,
					emp.department_id = department.id,
					emp.userId = usr.id,
			),emp.code+dept.code Code`
	str, _ = internal.Helper.QuoteExpression2(str)

	tk := sqlparser.NewStringTokenizer("select " + str)
	stm, err := sqlparser.ParseNext(tk)

	if err != nil {
		panic(err)
	}
	selectStm := stm.(*sqlparser.Select)
	for i := 0; i < b.N; i++ {
		smartier.from(selectStm)
	}

}
