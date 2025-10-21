package from

import (
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/sqlparser"
)

type DictionaryItem struct {
	Content string
	DbType  sqlparser.ValType
	Alias   string
}
type Dictionary struct {
	FieldMap map[string]DictionaryItem
	/*
				 Used to refer to an RDBMS table by its model name.

			Example:
				SELECT * FROM MyUser u
				// If the model "MyUser" refers to the table "my_users" in the RDBMS

			Result:
				Dictionary{
					AliasMap: {
						"myuser": "my_users",
						"u":      "my_users",
					}
				}

			When the compiler encounters a column with the qualifier "u" or "myuser",
			it will replace it with the actual table name "my_users".

			----

			Dùng để tham chiếu đến bảng trong RDBMS thông qua tên model.

		Ví dụ:
			SELECT * FROM MyUser u
			// Nếu model "MyUser" ánh xạ đến bảng "my_users" trong cơ sở dữ liệu

		Kết quả:
			Dictionary{
				AliasMap: {
					"myuser": "my_users",
					"u":      "my_users",
				}
			}

		Khi trình biên dịch gặp một cột có định danh (qualifier) là "u" hoặc "myuser",
		nó sẽ tự động thay thế bằng tên bảng thực tế "my_users".

	*/
	AliasMap        map[string]string
	AliasMapReverse map[string]string
	Entities        map[string]*entity.Entity
	/*
		Dung de tham chieu den tu table trong csdl den alias cua table o menh de join cua query

		Used to refer to a table in the database using the alias of the table in the join clause of the query.
	*/
	TableAlias map[string]string
}
