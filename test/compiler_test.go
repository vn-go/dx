package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/compiler"
	_ "github.com/vn-go/dx/compiler"
	_ "github.com/vn-go/dx/test/models" // registe all model
)

func TestCompilerSelectAllOneTable(t *testing.T) {
	ret, err := compiler.Compile("select * from User", "mysql")
	// expect := "`T1`.`id` `ID`,`T1`.`record_id` `RecordID`,`T1`.`created_at` `CreatedAt`,`T1`.`updated_at` `UpdatedAt`,`T1`.`user_id` `UserId`,`T1`.`email` `Email`,`T1`.`phone` `Phone`,`T1`.`username` `Username`,`T1`.`hash_password` `HashPassword`,`T1`.`description` `Description`"
	assert.NoError(t, err)
	t.Log(ret)
}
func BenchmarkCompilerSelectAllOneTable(t *testing.B) {
	for i := 0; i < t.N; i++ {
		compiler.Compile("select * from User", "mysql")
	}

}
func TestCompilerSelectAll2Table(t *testing.T) {
	ret, err := compiler.Compile("select concat(u.userId,?,department.code) as FName, u.username  from user u,department", "mysql")
	// expect := "`T1`.`id` `ID`,`T1`.`record_id` `RecordID`,`T1`.`created_at` `CreatedAt`,`T1`.`updated_at` `UpdatedAt`,`T1`.`user_id` `UserId`,`T1`.`email` `Email`,`T1`.`phone` `Phone`,`T1`.`username` `Username`,`T1`.`hash_password` `HashPassword`,`T1`.`description` `Description`"
	assert.NoError(t, err)
	t.Log(ret)
}

func TestCompilerSelect1Table(t *testing.T) {
	ret, err := compiler.Compile("select User.*,Department.* from User,Department", "mysql")
	// expect := "`T1`.`id` `ID`,`T1`.`record_id` `RecordID`,`T1`.`created_at` `CreatedAt`,`T1`.`updated_at` `UpdatedAt`,`T1`.`user_id` `UserId`,`T1`.`email` `Email`,`T1`.`phone` `Phone`,`T1`.`username` `Username`,`T1`.`hash_password` `HashPassword`,`T1`.`description` `Description`"
	assert.NoError(t, err)
	t.Log(ret)
}
func TestSelectFormWhere1(t *testing.T) {
	ret, err := compiler.Compile(`
								select 
									user.id,user.name 
									from user u  inner join Department d on u.id=d.userId
									where len(concat(user.Code,?))>1`,
		"mysql")
	// expect := "`T1`.`id` `ID`,`T1`.`record_id` `RecordID`,`T1`.`created_at` `CreatedAt`,`T1`.`updated_at` `UpdatedAt`,`T1`.`user_id` `UserId`,`T1`.`email` `Email`,`T1`.`phone` `Phone`,`T1`.`username` `Username`,`T1`.`hash_password` `HashPassword`,`T1`.`description` `Description`"
	assert.NoError(t, err)
	t.Log(ret)
}
func TestCompileSelect(t *testing.T) {
	ret, err := compiler.CompileSelect("user.userId", "mysql")
	assert.NoError(t, err)
	t.Log(ret)
}
func BenchmarkCompileSelect(t *testing.B) {
	for i := 0; i < t.N; i++ {
		compiler.CompileSelect("user.userId", "mysql")

	}

}
