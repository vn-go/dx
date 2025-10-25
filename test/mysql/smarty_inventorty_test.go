package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestSmartyInventory(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	expetedAccessScope := `{
  "department.id": {
    "EntityName": "Department",
    "EntityFieldName": "ID"
  },
  "department.name": {
    "EntityName": "Department",
    "EntityFieldName": "Name"
  },
  "department.parentid": {
    "EntityName": "Department",
    "EntityFieldName": "ParentID"
  },
  "user.departmentid": {
    "EntityName": "User",
    "EntityFieldName": "DepartmentId"
  },
  "user.id": {
    "EntityName": "User",
    "EntityFieldName": "Id"
  },
  "user.username": {
    "EntityName": "User",
    "EntityFieldName": "Username"
  }
}`
	expectedSql := "SELECT `u`.`id` `Id`, `u`.`username` `Username`, `depTree`.`Name` `departmentName` FROM `sys_users` `u` join  (SELECT `dep`.`id` `ID`, `dep`.`parent_id` `ParentID`, count(`child`.`id`) `ChildCount`, `dep`.`name` `Name`, `child`.`name` `ChildName` FROM `departments` `dep` left join  `departments` `child` ON `dep`.`id` = `child`.`parent_id` GROUP BY `child`.`name`, `dep`.`id`, `dep`.`name`, `dep`.`parent_id`) `deptree` ON `u`.`department_id` = `depTree`.`id`"

	sql, err := db.Smart(`
		/* sub query self join department*/
		subsets(
						dep.id, dep.parentId, count(child.id) ChildCount, 
						dep.name, child.name ChildName, 

						from(
									department dep, 
									department child ,
									left(dep.id = child.parentId)
						)
				) 			 depTree,
		/*end of subset */
		u.id,
		
		u.username, 
		depTree.Name departmentName, 
		from(
			user u, 
			
			depTree,
			u.departmentId = depTree.id 
			)

	`)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, expectedSql, sql.Query)
	assert.Equal(t, expetedAccessScope, sql.ScopeAccess.String())
}
func BenchmarkSmartyInventory(b *testing.B) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		b.Error(err)
	}
	defer db.Close()
	inputQuery := `
		/* sub query self join department*/
		subsets(
						dep.id, dep.parentId, count(child.id) ChildCount, 
						dep.name, child.name ChildName, 

						from(
									department dep, 
									department child ,
									left(dep.id = child.parentId)
						)
				) 			 depTree,
		/*end of subset */
		u.id,
		
		u.username, 
		depTree.Name departmentName, 
		from(
			user u, 
			
			depTree,
			u.departmentId = depTree.id 
			)

	`
	expetedAccessScope := `{
  "department.id": {
    "EntityName": "Department",
    "EntityFieldName": "ID"
  },
  "department.name": {
    "EntityName": "Department",
    "EntityFieldName": "Name"
  },
  "department.parentid": {
    "EntityName": "Department",
    "EntityFieldName": "ParentID"
  },
  "user.departmentid": {
    "EntityName": "User",
    "EntityFieldName": "DepartmentId"
  },
  "user.id": {
    "EntityName": "User",
    "EntityFieldName": "Id"
  },
  "user.username": {
    "EntityName": "User",
    "EntityFieldName": "Username"
  }
}`
	expectedSql := "SELECT `u`.`id` `Id`, `u`.`username` `Username`, `depTree`.`Name` `departmentName` FROM `sys_users` `u` join  (SELECT `dep`.`id` `ID`, `dep`.`parent_id` `ParentID`, count(`child`.`id`) `ChildCount`, `dep`.`name` `Name`, `child`.`name` `ChildName` FROM `departments` `dep` left join  `departments` `child` ON `dep`.`id` = `child`.`parent_id` GROUP BY `child`.`name`, `dep`.`id`, `dep`.`name`, `dep`.`parent_id`) `deptree` ON `u`.`department_id` = `depTree`.`id`"
	b.Run("SmartyInventory", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sql, err := db.Smart(inputQuery)
			if err != nil {
				panic(err)
			}

			assert.Equal(b, expectedSql, sql.Query)
			assert.Equal(b, expetedAccessScope, sql.ScopeAccess.String())
		}
	})
	b.Run("SmartyInventoryParallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sql, err := db.Smart(inputQuery)
				if err != nil {
					panic(err)
				}

				assert.Equal(b, expectedSql, sql.Query)
				assert.Equal(b, expetedAccessScope, sql.ScopeAccess.String())
			}
		})
	})

}
