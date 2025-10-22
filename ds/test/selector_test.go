package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/ds"
)

func TestSimpleSelect(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	sql, err := ds.Compile(

		dialect,
		`
			select(item(id, name, price))
		`,
	)
	if err != nil {
		t.Error(err)
	}
	expectesSql := "SELECT [items].[id] [ID], [items].[name] [Name], [items].[price] [Price] FROM [items]"

	assert.Equal(t, expectesSql, sql.Sql)
	fmt.Println(sql.Sql)
}
func TestSimpleSelectAndWhere(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	sql, err := ds.Compile(

		dialect,
		`
			select(item(id, name, price))
			where(id = 1)
		`,
	)
	if err != nil {
		t.Error(err)
	}
	expectesSql := "SELECT [items].[id] [ID], [items].[name] [Name], [items].[price] [Price] FROM [items]"

	assert.Equal(t, expectesSql, sql.Sql)
	fmt.Println(sql.Sql)
}
func TestSimpleSelectAndWhereIsSelect(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	sql, err := ds.Compile(

		dialect,
		`
			select(item(id, name, price))
			where(
					id = get(select(id) from(item(id)) where(id = 1))
				)
		`,
	)
	if err != nil {
		t.Error(err)
	}
	expectesSql := "SELECT [items].[id] [ID], [items].[name] [Name], [items].[price] [Price] FROM [items]"

	assert.Equal(t, expectesSql, sql.Sql)
	fmt.Println(sql.Sql)
}
func BenchmarkSimpleSelect(b *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(b, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	for i := 0; i < b.N; i++ {
		_, err = ds.Compile(

			dialect,
			`
			select(item(id, name, price))
		`,
		)
		if err != nil {
			b.Fail()
		}
	}

}
func TestSimpleSelectuUnion(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	sql, err := ds.Compile(

		dialect,
		`	from(
				imventory(
					select(IncrementDetail(itemId, amount))
					union
					select(DecrementDetail(itemId, amount*-1))
				
				)
			)
			select(
				imventory.itemId, sum(imventory.amount) amount	
			)
			
		`,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sql.Sql)
	expectesSql := "SELECT [imventory].[ItemID], sum([imventory].[Amount]) FROM SELECT [increment_details].[item_id] [ItemID], [increment_details].[amount] [Amount] FROM [increment_details]\n union \n SELECT [decrement_details].[item_id] [ItemID], [decrement_details].[amount] * @p1 [] FROM [decrement_details]"

	assert.Equal(t, expectesSql, sql.Sql)

}
func BenchmarkSimpleSelectuUnion(t *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	for i := 0; i < t.N; i++ {
		_, err = ds.Compile(

			dialect,
			`
				select(incrementDetail(id, amount))
				union
				select(decrementDetail(id, amount))
			`,
		)
		if err != nil {
			t.Fail()
		}

	}

}
