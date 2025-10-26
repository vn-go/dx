package mysql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestSmartyUnion(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	expectedSql := " SELECT `i`.`name` `Name`, `i`.`price` `Price`, `d`.`amount` `Amount` FROM `items` `i` join  `increment_details` `d` ON `i`.`id` = `d`.`item_id` union all SELECT `i`.`name` `Name`, `i`.`price` `Price`, `d`.`amount` `Amount` FROM `items` `i` join  `decrement_details` `d` ON `i`.`id` = `d`.`item_id`"
	sql, err := db.Smart(`
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, incrementDetail d)) inc,
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, decrementDetail d)) dec,
		union(inc+dec)

	`)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedSql, strings.ReplaceAll(sql.Query, "\n", ""))
	fmt.Print(sql.ScopeAccess.String())
}

func BenchmarkSmartyUnion(t *testing.B) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	query := `
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, incrementDetail d)) inc,
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, decrementDetail d)) dec,
		union(inc+dec)

	`
	expectedSql := " SELECT `i`.`name` `Name`, `i`.`price` `Price`, `d`.`amount` `Amount` FROM `items` `i` join  `increment_details` `d` ON `i`.`id` = `d`.`item_id` union all SELECT `i`.`name` `Name`, `i`.`price` `Price`, `d`.`amount` `Amount` FROM `items` `i` join  `decrement_details` `d` ON `i`.`id` = `d`.`item_id`"
	expecdAccessScope := `{
  "decrementdetail.amount": {
    "EntityName": "DecrementDetail",
    "EntityFieldName": "Amount"
  },
  "decrementdetail.itemid": {
    "EntityName": "DecrementDetail",
    "EntityFieldName": "ItemID"
  },
  "incrementdetail.amount": {
    "EntityName": "IncrementDetail",
    "EntityFieldName": "Amount"
  },
  "incrementdetail.itemid": {
    "EntityName": "IncrementDetail",
    "EntityFieldName": "ItemID"
  },
  "item.id": {
    "EntityName": "Item",
    "EntityFieldName": "ID"
  },
  "item.name": {
    "EntityName": "Item",
    "EntityFieldName": "Name"
  },
  "item.price": {
    "EntityName": "Item",
    "EntityFieldName": "Price"
  }
}`
	t.Run("SmartyUnion", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {

			sql, err := db.Smart(query)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, expectedSql, strings.ReplaceAll(sql.Query, "\n", ""))
			assert.Equal(b, expecdAccessScope, sql.ScopeAccess.String())
		}
	})
	t.Run("SmartyUnionParallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sql, err := db.Smart(query)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, expectedSql, strings.ReplaceAll(sql.Query, "\n", ""))
				assert.Equal(b, expecdAccessScope, sql.ScopeAccess.String())
			}
		})
	})

}
func TestSmartyUnion2(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	expectedSql := " SELECT `i`.`name` `Name`, `i`.`price` `Price`, `d`.`amount` `Amount` FROM `items` `i` join  `increment_details` `d` ON `i`.`id` = `d`.`item_id` union all SELECT `i`.`name` `Name`, `i`.`price` `Price`, `d`.`amount` `Amount` FROM `items` `i` join  `decrement_details` `d` ON `i`.`id` = `d`.`item_id`"
	sql, err := db.Smart(`
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, incrementDetail d)) inc,
		subsets(i.name,i.price,d.amount, from( i.id=d.itemId, item i, decrementDetail d)) dec,
		subsets(union(inc+dec)) all,
		from( all), all.name,sum(all.amount) TotalAmount

	`)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedSql, strings.ReplaceAll(sql.Query, "\n", ""))
	fmt.Print(sql.ScopeAccess.String())
}
