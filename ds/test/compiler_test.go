package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/ds"

	_ "github.com/vn-go/dx/ds/models"
)

var cnn = "sqlserver://sa:123456@localhost:1433?database=hrm"

func TestParseClausesFromInSelect(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
				select(item(id,name) )  /* -> select id,name from item*/
					
		`

	sql, err := ds.Compile(

		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
func TestParseClausesFromInSelect1(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
				select(item(id,name) incData )  /* -> select id,name from item*/
					
		`

	sql, err := ds.Compile(

		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
func TestParseClausesFromInSelec1t2(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
				select(
					item(id ID,name Name) incData,
					incrementDetail(amount) incDetail, 
					incData.id=incDetail.itemId
				)  /* -> select id,name from item*/
					
		`

	sql, err := ds.Compile(
		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
func TestParseClauses0(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
					from(incrementDetail)
					select(id,sum(amount) TotalAmount)
					where(id > 30)
					orderby(id desc)
		`

	sql, err := ds.Compile(

		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
func TestParseClauses0_1(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
						from(incrementDetail)
						select(id,sum(amount) TotalAmount)
						where(id > 30)
						union all
						from(decrementDetail)
						select(id,sum(amount) TotalAmount)
						where(id > 30)
		`

	sql, err := ds.Compile(

		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
func TestParseClauses1(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
					from(
						qr1 (
							from(incrementDetail)
							select(id,sum(amount) TotalAmount)
							where(id > 30)
							
						) 
						
					) 
					select(TotalAmount)
					where(TotalAmount > 30)
					orderby(TotalAmount desc)
		`

	sql, err := ds.Compile(

		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
func TestParseClauses2(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `	from(
						qr1 (
							from(incrementDetail)
							select(id,sum(amount) TotalAmount)
							where(id > 30)
						union all
							from(decrementDetail)
							select(id,sum(amount) TotalAmount)
							where(id > 30)
						
						)
					)
				select(TotalAmount)
						
		`

	sql, err := ds.Compile(

		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
func TestParseClauses3(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `	from( 	
						increment INC,
						increment.id=incrementDetail.incrementId
						/*left(increment.id=incrementDetail.incrementId),
						leftOuter(incrementDetail.itemId=item.id),*/
					)
				select(TotalAmount)
						
		`

	sql, err := ds.Compile(

		dialect,
		query,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
}
