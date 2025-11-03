package compact_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
)

var dsn = "postgres://postgres:123456@localhost:5432/hrm?sslmode=disable&"

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
