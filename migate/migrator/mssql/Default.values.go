package mssql

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

/*
Hàm này sẽ trả về một map chứa các giá trị mặc định của
các trường dữ liệu được định nghĩa trong struct.
Ví dụ:

	type User struct {
		CreatedAt time.Time `db:"default:now"` //<-- khi chạy với mssql,
		// now sẽ chuyển thành "GETDATE()"
	}
		Vì vậy hệ thống cần map định nghĩa các giá trị mặc định của các trường dữ liệu
		được định nghĩa trong struct với các giá trị hoặc hàm được hỗ trợ bởi RDMMS.
		Trong trường hợp này, hàm GetGetDefaultValueByFromDbTag() sẽ trả về một map chứa
		các giá trị mặc định mà được khai báo trong tag `default` của field trong struct
		với MSSQL.
	---------------------------
	This function will return a map containing the default values of the data fields defined in a struct.

For example:

	type User struct {
	    CreatedAt time.Time `db:"default:now"` // <-- when running with MSSQL,
	    // "now" will be converted to "GETDATE()"
	}
		Therefore, the system needs to map the default values of the data fields defined in the struct to the values or functions supported by the RDBMS.

In this case, the function GetGetDefaultValueByFromDbTag() will return a map containing the default values declared in the default tag of the struct fields for MSSQL.
*/
func (m *migratorMssql) GetGetDefaultValueByFromDbTag() map[string]string {
	return map[string]string{
		"now()":  "GETDATE()", //<-- chỗ này sẽ mở rộng cho các case cụ thể về sau
		"uuid()": "NEWID()",
	}
}

/*
Hàm này sẽ trả về một map chứa các kiểu dữ của GO có khả năng map với kiểu dữ của SQL Server.
Ví dụ:

	type User struct {
		ID        int       //<-- kiểu dữ của ID là int, DEV kg cần phải khai báo
							// `db:"type:int"` hệ thống sẽ tự động map kiểu dữ này với SQL Server

Vì vây, hệ thống cần map các kiểu dữ của GO với kiểu dữ của RDBMS cụ thể.
Trong trường hợp này cụ thể là SQL Server.

-----------------------------------

This function will return a map containing Go data types that can be mapped to SQL Server data types.
For example:

	type User struct {
	    ID int // <-- the type of ID is int, the developer does not need to declare
	           // `db:"type:int"` because the system will automatically map this type to SQL Server
	}

	Therefore, the system needs to map Go data types to the data types of a specific RDBMS.

In this particular case, it is SQL Server.
*/
func (m *migratorMssql) GetColumnDataTypeMapping() map[reflect.Type]string {
	return map[reflect.Type]string{
		reflect.TypeOf(""):          "nvarchar",
		reflect.TypeOf(int(0)):      "int",
		reflect.TypeOf(int8(0)):     "tinyint",
		reflect.TypeOf(int16(0)):    "smallint",
		reflect.TypeOf(int32(0)):    "int",
		reflect.TypeOf(int64(0)):    "bigint",
		reflect.TypeOf(uint(0)):     "bigint", // SQL Server doesn't support unsigned
		reflect.TypeOf(uint8(0)):    "tinyint",
		reflect.TypeOf(uint16(0)):   "int",
		reflect.TypeOf(uint32(0)):   "bigint",
		reflect.TypeOf(uint64(0)):   "bigint",
		reflect.TypeOf(float32(0)):  "real",
		reflect.TypeOf(float64(0)):  "float",
		reflect.TypeOf(bool(false)): "bit",
		reflect.TypeOf([]byte{}):    "varbinary",
		reflect.TypeOf(time.Time{}): "datetime2",
		reflect.TypeOf(uuid.UUID{}): "uniqueidentifier",
	}
}
