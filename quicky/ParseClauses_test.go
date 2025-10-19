package quicky

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/dialect/factory"
)

var cnn = "sqlserver://sa:123456@localhost:1433?database=hrm"

func TestParseClauses0(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	data, textParams, err := Clause.Inspect(`
	from(incrementDetail)
	select(id,sum(amount) TotalAmount)
	where(id > 30)
	orderby(id desc)
`)
	assert.NoError(t, err)
	t.Log(data)
	assert.NoError(t, err)
	t.Log(data)
	sqlParser := newSqlParser()
	dialect := factory.DialectFactory.Create(db.DriverName)
	sqlParser.Parse(data, dialect, textParams, 10000, 20000)
	fmt.Println(sqlParser.Statement)
}
func TestParseClauses1(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	data, textParams, err := Clause.Inspect(`
	from(user u,department d,department d2,u.id=d.id,d1.id=d2.id)
	select(name, age)
	where(age > 30)
	orderby(age desc)
`)
	assert.NoError(t, err)
	t.Log(data)
	sqlParser := newSqlParser()
	dialect := factory.DialectFactory.Create(db.DriverName)
	sqlParser.Parse(data, dialect, textParams, 10000, 20000)
	fmt.Println(sqlParser.Statement)
}
func TestParseClauses2(t *testing.T) {
	// fix requirement
	data, _, err := Clause.Inspect(`
	from(query(
		from(customers)
		select(name, age)
		where(age > 30)
	))
	select(name, age)
	where(age > 30)
	orderby(age desc)
	limit(10)
	skip(20)

`)
	assert.NoError(t, err)
	t.Log(data)
}
func TestParseClauses3(t *testing.T) {
	data, _, err := Clause.Inspect(`
	select(user(name Name, age), department(name), u.id=d.id)
	where(u.age > 30)
	orderby(age desc)
	limit(10)
	skip(20)
	
`)

	assert.NoError(t, err)
	t.Log(data)
}
func TestParseClauses4(t *testing.T) {
	data, _, err := Clause.Inspect(`
	select(user(name, age), department(name), left(u.id=d.id))
	where(u.age > 30)
	orderby(age desc)
	limit(10)
	skip(20)
	
`)

	assert.NoError(t, err)
	t.Log(data)
}
func TestParseClauses5(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	data, textParams, err := Clause.Inspect(`
	select(increment(name, age) i, decrement(name) d, i.id=d.id)
	where(u.age > 30)
	orderby(age desc)
	limit(10)
	skip(20)
	
`)

	assert.NoError(t, err)
	t.Log(data)
	sqlParser := newSqlParser()
	dialect := factory.DialectFactory.Create(db.DriverName)
	err = sqlParser.Parse(data, dialect, textParams, 10000, 20000)
	assert.NoError(t, err)
	fmt.Println(sqlParser.Statement)
}

func TestParseClauses6(t *testing.T) {
	data, _, err := Clause.Inspect(`
		
		select(user(name, age), department(name), u.id=d.id) where(u.age > 30))
		union all
		select(customer(name, age), department(name), u.id=d.id) where(u.age > 30))
`)
	assert.NoError(t, err)
	t.Log(data)
}
func TestParseClauses7(t *testing.T) {
	data, _, err := Clause.Inspect(`
		
		where(u.age > 30)) select(user(name, age,union), department(name), u.id=d.id) 
		union all
		select(customer(name, age), department(name), u.id=d.id) where(u.age > 30))
	`)
	assert.NoError(t, err)
	t.Log(data)
}
func TestParseClauses8(t *testing.T) {
	data, _, err := Clause.Inspect(`
		from(
			select(	
						increment(year(createdOn) day) inc, 
						incrementDetail(itemId, quantity) incDetail,
						items(name, price) itm,
						inc.id=incDetail.incrementId and incDetail.itemId=itm.id
					) 
			where(inc.day=12) 
					
			union all
					select(	
						decrement(year(createdOn) day) dec, 
						decrementDetail(itemId, quantity*(-1)) decDetail,
						items(name, price) itm,
						dec.id=decDetail.incrementId and decDetail.itemId=itm.id
					) 
			where (dec.day=12)
		),
		select(sum(quantity))
	`)
	if err != nil {
		fmt.Println(err)
	}
	assert.NoError(t, err)
	t.Log(data)
}
func TestParseClauses9(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()

	data, textParams, err := Clause.Inspect(`
		from(
			select(	
						increment(year(createdOn) day) inc,
						incrementDetail(itemId, quantity) incDetail,items(name, price) itm, 	/*voi bang incrementDetail chon cac cot */
						decrementDetail() decDetail, 											/*alias tablename*/
						decDetail.id+incDetail.id total,										/*khi muon truy cap 2 bang alias thi phai dung alias */
						left(inc.id=incDetail.incrementId) , 									/*left join */
						incDetail.itemId=itm.id,												/*join voi xuat hang va ban mat hang*/
						decDetail.id=incDetail.id
					) 
			where(inc.day=12 and no='a''bc') 
					
			union all
					select(	
						decrement(year(createdOn) day) dec, 
						decrementDetail(itemId, quantity*(-1)) 
						decDetail,items(name, price) itm,
						dec.id=decDetail.incrementId , decDetail.itemId=itm.id
					) 
			where (dec.day=12 and dec.day<?)
		),
		select(sum(quantity*?) Quantity)
		where(Quantity > 100)
	`)
	assert.NoError(t, err)
	t.Log(data)
	sqlParser := newSqlParser()
	dialect := factory.DialectFactory.Create(db.DriverName)
	sqlParser.Parse(data, dialect, textParams, 10000, 20000)
	fmt.Println(sqlParser.Statement)
}
