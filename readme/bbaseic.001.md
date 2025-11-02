# DSL Query – Hướng dẫn sử dụng

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
userInfos := []struct {
    Id       uint64 `db:"pk;auto" json:"id"`
    Username string `db:"size:50;uk" json:"username"`
}{}
```go
err = db.DslQuery(&userInfos, "user(id, username)")
if err != nil {
    panic(err)
}
## SELECT VỚI ĐIỀU KIỆN
----------------------------
SELECT `T1`.`id` `Id`, `T1`.`username` `Username` FROM `sys_users` `T1` WHERE `T1`.`username` = ?
----------------------------
```
userInfos := []struct {
    Id       uint64 `db:"pk;auto" json:"id"`
    Username string `db:"size:50;uk" json:"username"`
}{}
```go
err = db.DslQuery(&userInfos, "user(id, username),where(username='admin')")
if err != nil {
    panic(err)
}

## QUY TẮC BẮT BUỘC: **MỌI BIỂU THỨC KHÔNG LẤY TRỰC TIẾP TỪ CỘT PHẢI CÓ ALIAS**

> **BẮT LỖI NGAY TẠI GO – TRƯỚC KHI GỬI XUỐNG DATABASE ENGINE**

---

### Lỗi thường gặp

```go
err = db.DslQuery(&userInfos, "user(count(userid), roleid), where(username like '%%admin%%')")