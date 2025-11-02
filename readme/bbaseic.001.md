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
    ```