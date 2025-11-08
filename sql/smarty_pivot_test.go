package sql

import (
	"fmt"
	"testing"
)

func TestPivot001(t *testing.T) {
	query := `
				from(user u),
				crossTab(for(day(u.createdOn) Day ,1,31),select(count(u.id) Total, SUM(U.AMOUNT) AMOUNT))
				`
	sql, _, _, err := smartier.simple(query)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
}
