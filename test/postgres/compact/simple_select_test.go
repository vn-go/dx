package compact_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
	tModels "github.com/vn-go/dx/test/models"
)

var dsn = "postgres://postgres:123456@localhost:5432/hrm?sslmode=disable&"

func TestFindRole(t *testing.T) {
	dx.Options.ShowSql = true
	expectedData := `{"id":2302,"roleId":"e0d4573b-85cc-4b12-b146-a2f9eadc9d9f","code":"X-0000","name":"Role-0000","description":"","createdOn":"2025-10-09T18:04:28.523906+07:00","modifiedOn":null,"createdBy":"admin","isActive":true}`
	//dataResult := []byte(expectedData)
	db, err := dx.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	data, err := db.FindFirst(tModels.Role{}, "", "code='X-0000'")
	if err != nil {
		t.Fatal(err)
	}
	bff, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bff))
	assert.Equal(t, expectedData, string(bff))
}
func BenchmarkFindRole(b *testing.B) {
	expectedData := `{"id":2302,"roleId":"e0d4573b-85cc-4b12-b146-a2f9eadc9d9f","code":"X-0000","name":"Role-0000","description":"","createdOn":"2025-10-09T18:04:28.523906+07:00","modifiedOn":null,"createdBy":"admin","isActive":true}`
	//dataResult := []byte(expectedData)
	db, err := dx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	b.Run("FindRole", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, err := db.FindFirst(tModels.Role{}, "", "code='X-0000'")
			if err != nil {
				panic(err)
			}
			bff, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, expectedData, string(bff))
		}
	})
	b.Run("FindRole-Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				data, err := db.FindFirst(tModels.Role{}, "", "code='X-0000'")
				if err != nil {
					panic(err)
				}
				bff, err := json.Marshal(data)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, expectedData, string(bff))
			}
		})
	})

}
func TestSimpleSelect(t *testing.T) {
	db, err := dx.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	//user:=models.Customers{}
	sql, err := db.ParseDsl("user(username+email fullName) ,where(user.username+user.email='admin')")
	if err != nil {
		panic(err)
	}
	expectSql := `SELECT "T1"."id" "Id", "T1"."user_id" "UserId", "T1"."username" "Username", "T1"."hash_password" "HashPassword", "T1"."email" "Email", "T1"."phone" "Phone", "T1"."created_on" "CreatedOn", "T1"."modified_on" "ModifiedOn", "T1"."is_active" "IsActive", "T1"."latest_login_fail" "LatestLoginFail", "T1"."latest_login" "LatestLogin", "T1"."role_code" "RoleCode", "T1"."last_time_change_password" "LastTimeChangePassword", "T1"."is_tenant_admin" "IsTenantAdmin", "T1"."role_id" "RoleId", "T1"."is_sys_admin" "IsSysAdmin", "T1"."created_by" "CreatedBy", "T1"."department_id" "DepartmentId" FROM "sys_users" "T1"`
	assert.Equal(t, expectSql, sql.Query)
	fmt.Println(sql.Query)
	fmt.Println(sql.OutputFields.String())
}
func TestQuery(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := db.QueryModel(tModels.User{}).RightJoin(tModels.Department{}, "user.departmentId=department.id").Select(
		"user()",
	)
	r.Limit(10)
	r.Offset(20)
	r.Sort("Id desc")
	sql, err := r.Analize()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)
	fmt.Println(sql.OutputFields.String())
	total, err := r.Count()
	if err != nil {
		panic(err)
	}
	fmt.Println(total)
}
func BenchmarkQuery(b *testing.B) {
	db, err := dx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sqlExpected := `SELECT "user"."id" "Id", "user"."user_id" "UserId", "user"."username" "Username", "user"."hash_password" "HashPassword", "user"."email" "Email", "user"."phone" "Phone", "user"."created_on" "CreatedOn", "user"."modified_on" "ModifiedOn", "user"."is_active" "IsActive", "user"."latest_login_fail" "LatestLoginFail", "user"."latest_login" "LatestLogin", "user"."role_code" "RoleCode", "user"."last_time_change_password" "LastTimeChangePassword", "user"."is_tenant_admin" "IsTenantAdmin", "user"."role_id" "RoleId", "user"."is_sys_admin" "IsSysAdmin", "user"."created_by" "CreatedBy", "user"."department_id" "DepartmentId" FROM "sys_users" "user" ORDER BY "user"."id" desc LIMIT $2 OFFSET $1`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := db.QueryModel(tModels.User{})
		r.Limit(10)
		r.Offset(20)
		r.Sort("id desc")
		sql, err := r.Analize()
		if err != nil {
			panic(err)
		}
		assert.Equal(b, sqlExpected, sql.Query)

		total, err := r.Count()
		if err != nil {
			panic(err)
		}
		assert.Equal(b, int64(0), total)
		// fmt.Println(total)
		// fmt.Println(sql.Query)
		//fmt.Println(sql.OutputFields.String())
	}
}
