package sql

import (
	"fmt"
	"testing"
)

func TestSmartyUnion(t *testing.T) {
	str := `
	subSets(employee.code,where(employee.id = 1)) qr1,
	subSets(user.username,where(user.id = 2)) qr2,
	subSets(user.username,where(user.id = 3)) qr3,
	union(qr1+qr2*qr3),
	`

	sql, _, err :=smartier.simple(str)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
}
func TestSmartyUnion2(t *testing.T) {
	str := `
	subSets(employee.code,where(employee.id = 1)) qr1,
	subSets(user.username,where(user.id = 2)) qr2,
	subSets(user.username,where(user.id = 3)) qr3,
	subSets(union(qr1+qr2*qr3)) qr4,
	from(qr4, qr1 , qr4.username = qr1.code, qr2, qr4.username = qr2.username), qr4.code
	`
	str = `
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, incrementDetail d)) inc,
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, decrementDetail d)) dec,
		subsets(union(inc+dec)) all,
		from( all), all.name,sum(all.amount) TotalAmount

	`
	sql, _, err :=smartier.simple(str)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
}
