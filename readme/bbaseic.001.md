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
----------------------------
SELECT `T1`.`id` `Id`, `T1`.`username` `Username` FROM `sys_users` `T1`
----------------------------
```

