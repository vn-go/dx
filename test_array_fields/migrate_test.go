package testarrayfields

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/vn-go/dx"
)

var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"

func Test0001(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	dep, err := dx.NewDTO[Department]()
	if err != nil {
		panic(err)
	}
	dep.Code = "A001"
	dep.Name = "XXXX"
	dep.RecordID = uuid.NewString()
	dep.ChildrenID = []int{1, 2, 3}
	err = db.Insert(dep)
	if err != nil {
		panic(err)
	}
}
func TestSmartyWithArrayFilter(t *testing.T) {
	//JSON_OVERLAPS(children_id, JSON_ARRAY(2,4))
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sql, err := db.Smart(`
		dep(count(id+?) Count),
		sum(dep.id),
		from(
			department child,
			department dep,
			list.contains(dep.ChildrenId,child.ChildrenId)
		),
		
	`, 1, []int{1, 2})
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)
}
func TestSmartyRowset(t *testing.T) {
	//JSON_OVERLAPS(children_id, JSON_ARRAY(2,4))
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sql, err := db.Smart(`
		rowset(user(departmentId),where)
		dataset.department(count(id) Count),
		where(id=)
	`, []int{1, 2})
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)
}
