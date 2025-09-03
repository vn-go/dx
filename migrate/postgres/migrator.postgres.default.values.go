package postgres

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

func (m *MigratorPostgres) GetGetDefaultValueByFromDbTag() map[string]string {
	return map[string]string{
		"now()":              "CURRENT_TIMESTAMP", // hoặc "now()" đều hợp lệ với Postgres
		"uuid_generate_v4()": "uuid_generate_v4()",
		"uuid":               "uuid_generate_v4()",
	}
}
func (m *MigratorPostgres) GetColumnDataTypeMapping() map[reflect.Type]string {
	return map[reflect.Type]string{
		reflect.TypeOf(""):          "citext",
		reflect.TypeOf(int(0)):      "integer",
		reflect.TypeOf(int8(0)):     "smallint",
		reflect.TypeOf(int16(0)):    "smallint",
		reflect.TypeOf(int32(0)):    "integer",
		reflect.TypeOf(int64(0)):    "bigint",
		reflect.TypeOf(uint(0)):     "bigint", // PostgreSQL không có unsigned, nên dùng type đủ lớn
		reflect.TypeOf(uint8(0)):    "smallint",
		reflect.TypeOf(uint16(0)):   "integer",
		reflect.TypeOf(uint32(0)):   "bigint",
		reflect.TypeOf(uint64(0)):   "numeric", // an toàn cho số lớn
		reflect.TypeOf(float32(0)):  "real",
		reflect.TypeOf(float64(0)):  "double precision",
		reflect.TypeOf(bool(false)): "boolean",
		reflect.TypeOf([]byte{}):    "bytea",
		reflect.TypeOf(time.Time{}): "timestamp with time zone",
		reflect.TypeOf(uuid.UUID{}): "uuid",
	}
}
