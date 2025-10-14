package oracle

import (
	"fmt"
	"reflect"

	"github.com/vn-go/dx/entity"
)

// ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_chk_name_length;
func (m *MigratorOracle) createCheckLenConstraint1(tableName string, col entity.ColumnDef) string {
	if col.Field.Type == reflect.TypeFor[string]() || col.Field.Type == reflect.TypeFor[*string]() {
		checkSyntax := fmt.Sprintf("CHECK (char_length(%s) <= %d)", m.Quote(col.Name), *col.Length)

		constraintCheckName := fmt.Sprintf("%s_chk_%s_length", tableName, col.Name)
		dropConstraint := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;", m.Quote(tableName), m.Quote(constraintCheckName))
		sqlCreateCheckLen := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", m.Quote(tableName), m.Quote(constraintCheckName), checkSyntax)
		return dropConstraint + sqlCreateCheckLen
	}
	return ""

}
func (m *MigratorOracle) createCheckLenConstraint(tableName string, col entity.ColumnDef) string {
	if col.Field.Type == reflect.TypeFor[string]() || col.Field.Type == reflect.TypeFor[*string]() {
		checkSyntax := fmt.Sprintf("CHECK (char_length(%s) <= %d)", m.Quote(col.Name), *col.Length)

		constraintCheckName := fmt.Sprintf("%s_chk_%s_length", tableName, col.Name)

		sqlCreateCheckLen := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", m.Quote(tableName), m.Quote(constraintCheckName), checkSyntax)
		retSql := fmt.Sprintf(`
					DO $$
					BEGIN
						IF NOT EXISTS (
							SELECT 1
							FROM pg_constraint
							WHERE conname = '%s'
							AND conrelid = '%s'::regclass
						) THEN
							%s;
						END IF;
					END$$;
					`, constraintCheckName, tableName, sqlCreateCheckLen)
		return retSql
	}
	return ""

}
