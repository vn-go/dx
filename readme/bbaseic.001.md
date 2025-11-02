# DSL Query – Hướng dẫn sử dụng
## ALIAS TRONG DSL – CÁCH HOẠT ĐỘNG

| Nguồn | Ưu tiên | Ví dụ |
|------|--------|------|
| `as tên` trong DSL | 1 | `location(city as cityName)` → `cityName` |
| Tên field trong struct | 2 | `City string` → `City` |
| Tên cột DB | 3 | `city` → `city` |

> **Lưu ý:** Nếu struct có field `City`, thì `as cityName` → **bị bỏ qua**
> Viết truy vấn SQL ngắn gọn, dễ đọc, tự động xử lý `JOIN`, `UNION`, `GROUP BY`, `COUNT`, `SUM`, v.v.

---

## CÀI ĐẶT & KẾT NỐI MYSQL

```go
var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"
dx.Options.ShowSql = true
db, err := dx.Open("mysql", dsn)
if err != nil {
    panic(err)
}
## SELECT CƠ BẢN – LẤY CÁC CỘT CỤ THỂ
----------------------------
SELECT `T1`.`id` `Id`, `T1`.`username` `Username` FROM `sys_users` `T1`
----------------------------
```

```go
userInfos := []struct {
    Id       uint64 `db:"pk;auto" json:"id"`
    Username string `db:"size:50;uk" json:"username"`
}{}
err = db.DslQuery(&userInfos, "user(id, username)")
if err != nil {
    panic(err)
}
## SELECT VỚI ĐIỀU KIỆN
----------------------------
SELECT `T1`.`id` `Id`, `T1`.`username` `Username` FROM `sys_users` `T1` WHERE `T1`.`username` = ?
----------------------------
```

```go
userInfos := []struct {
    Id       uint64 `db:"pk;auto" json:"id"`
    Username string `db:"size:50;uk" json:"username"`
}{}

err = db.DslQuery(&userInfos, "user(id, username),where(username='admin')")
if err != nil {
    panic(err)
}

## QUY TẮC BẮT BUỘC: **MỌI BIỂU THỨC KHÔNG LẤY TRỰC TIẾP TỪ CỘT PHẢI CÓ ALIAS**

> **BẮT LỖI NGAY TẠI GO – TRƯỚC KHI GỬI XUỐNG DATABASE ENGINE**

```

### Lỗi thường gặp

```go
err = db.DslQuery(&userInfos, "user(count(userid), roleid), where(username like '%admin%')")

------------
panic: Please add a name (alias) for the expression 'count(user.userid)'. [recovered, repanicked]

---------------------------
```
## VÍ DỤ: JOIN 3 BẢNG – `user` → `role` → `department`

> **Mục tiêu:**  
> Lấy thông tin người dùng (`id`, `username`) cùng với **tên vai trò** và **tên phòng ban**.

---

### DSL Query – CHỈ 7 DÒNG

```go
query := `
    user(id, username), 
    role(name as roleName), 
    department(name as deptName),
    
    from(
        left(user.roleId = role.id), 
        left(user.departmentId = department.id)
    ),
    
    where(user.id = ?)
`
```

```go
type Location struct {
		// Khóa chính (BẮT BUỘC)
		ID int `db:"pk;auto" json:"id"`

		// Các cột khác
		City string `db:"size:100" json:"city"`
		Code string `db:"size:50" json:"code"`

		// ... các trường cần thiết khác
		// BaseModel (Nếu bạn sử dụng BaseModel)
	}
	dx.AddModels(&Location{}) // add model to dx truoc khi open database
	var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	userInfos := []struct {
		Users  uint64 `db:"pk;auto" json:"id"`
		RoleId string `db:"size:50;uk" json:"username"`
	}{}

	query := `user(id, username), 
          role(name as roleName), 
          department(name as deptName),
          location(city), //-- Thêm bảng thứ 4
          
          from(left(user.roleId = role.id), 
               left(user.departmentId = department.id), 
               left(department.locationId = location.id)), //-- Thêm điều kiện nối thứ 3

          where(user.id = ?)`
	err = db.DslQuery(&userInfos, query, 10, 100)
	if err != nil {
		panic(err)
	}
----------------------------
SELECT `T1`.`id` `Id`, `T1`.`username` `Username`, `T2`.`name` `roleName`, `T3`.`name` `deptName`, `T4`.`city` `City` FROM `sys_users` `T1` left join  `sys_roles` `T2` ON `T1`.`role_id` = `T2`.`id` join  `departments` `T3` ON `T1`.`department_id` = `T3`.`id` join  `locations` `T4` ON `T3`.`location_id` = `T4`.`id` WHERE `T1`.`id` = ?
----------------------------

```
# Hướng dẫn sử dụng DSQL (DSL Query) với `dx` trong Go

Hướng dẫn này giúp bạn **kết hợp dữ liệu từ nhiều bảng khác nhau** (`Account`, `Admin`, `Manager`) vào **một slice struct duy nhất** bằng cách sử dụng **cú pháp DSQL** (Domain Specific Language Query) của thư viện `dx`.

---

## 1. Cấu trúc Model (Struct)

```go
type Account struct {
    ID       uint64 `db:"pk;auto" json:"id"`
    Name     string `db:"size:100" json:"name"`
    Email    string `db:"size:100" json:"email"`
    IsActive bool   `db:"default:true" json:"isActive"`
}

type Admin struct {
    ID    uint64 `db:"pk;auto" json:"id"`
    Name  string `db:"size:100" json:"name"`
    Role  string `db:"size:50" json:"role"`
    Level int    `db:"default:1" json:"level"`
}

type Manager struct {
    ID         uint64  `db:"pk;auto" json:"id"`
    Name       string  `db:"size:100" json:"name"`
    Department string  `db:"size:100" json:"department"`
    Salary     float64 `db:"default:0" json:"salary"`
}

dx.AddModels(&Account{}, Admin{}, Manager{}) // add model to dx truoc khi open database
	var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	people := []struct {
		Id   uint64 `db:"pk;auto" json:"id"`
		Name string `db:"size:50;uk" json:"username"`
	}{}

	query := `
		   account(id, name)+
			admin(id, name)+
			manager(id, name)
	  `
	err = db.DslQuery(&people, query)
	if err != nil {
		panic(err)
	}

----------------------------

 
 SELECT `T1`.`id` `ID`, `T1`.`name` `Name` FROM `accounts` `T1`
 union all
 SELECT `T1`.`id` `ID`, `T1`.`name` `Name` FROM `admins` `T1`
 union all
 SELECT `T1`.`id` `ID`, `T1`.`name` `Name` FROM `managers` `T1`
----------------------------

```