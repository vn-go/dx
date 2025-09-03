package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	dialectCommon "github.com/vn-go/dx/dialect/common"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/errors"
	migrator "github.com/vn-go/dx/migrate"
	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
	"github.com/vn-go/xdb/migrate"
)

type entityInfo struct {
	tableName string
	entity    *common.Entity
}
type inserter struct {
}

func (r inserter) getEntityInfo(typ reflect.Type) (*entityInfo, error) {
	model := common.ModelRegistry.GetModelByType(typ)
	if model == nil {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		return nil, errors.NewModelError(typ)
	}
	tableName := model.GetTableName()
	entity := model.GetEntity()
	return &entityInfo{
		tableName: tableName,
		entity:    &entity,
	}, nil
}

func (r inserter) fetchAfterInsert(res sql.Result, entity *common.Entity, dataValue reflect.Value) error {
	// Nếu không có cột tự tăng thì bỏ qua
	autoCols := entity.GetAutoValueColumns()
	if len(autoCols) == 0 {
		return nil
	}

	lastID, err := res.LastInsertId() //<--loi khi chay voi mssql
	/*
		Cau query thuc the duoc goi den mssql nhu co dang sau
		 vi du
		 "LastInsertId is not supported. Please use the OUTPUT clause or add `select ID = convert(bigint, SCOPE_IDENTITY())` to the end of your query"  khi chay cau query sau
		 "INSERT INTO [departments] (
		 	[name],
			[code],
			[parent_id], ...) OUTPUT INSERTED.[id] VALUES (@p1, @p2, @p3, ...)"
	*/
	if err != nil {
		return err
	}

	for _, col := range autoCols {
		valField := dataValue.FieldByName(col.Field.Name)

		if valField.CanConvert(col.Field.Type) {
			valField.Set(reflect.ValueOf(lastID).Convert(col.Field.Type))

		}
	}

	return nil
}

/*
Rac roi o cho muon lay gia tri cua cac cot tự tăng khi insert trong sql server
phai dung db.Queryrow(sql, args)
Vi vay ham nay chi dung voi sqlserver
------------------------
There is confusion about how to get the value of auto-increment columns when inserting in SQL Server.
You must use db.QueryRow(sql, args).
Therefore, this function only works with SQL Server.
*/
func fetchAfterInsertForQueryRow(entity *common.Entity, dataValue reflect.Value, insertedValue any) error {
	autoCols := entity.GetAutoValueColumns()
	if len(autoCols) == 0 || insertedValue == nil {
		return nil
	}
	if len(autoCols) != 1 {
		return fmt.Errorf("only single auto-increment column is supported for QueryRow insert")
	}

	col := autoCols[0]
	field := dataValue.FieldByName(col.Field.Name)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("cannot set field %s", col.Field.Name)
	}

	val := reflect.ValueOf(insertedValue)
	// Nếu là ptr như *int64
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	// Gán giá trị nếu kiểu phù hợp
	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
	} else if val.Type().ConvertibleTo(field.Type()) {
		field.Set(val.Convert(field.Type()))
	} else {
		return fmt.Errorf("cannot assign insert id type %s to field %s type %s",
			val.Type(), col.Field.Name, field.Type())
	}

	return nil
}

func (r *inserter) InsertWithTx(tx *tenantDB.TenantTx, data interface{}) error {
	dialect := factory.DialectFactory.Create(tx.Db.GetDriverName())
	dataValue := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		dataValue = dataValue.Elem()
	}
	repoType, err := r.getEntityInfo(typ)
	if err != nil {
		return err
	}
	sql, args := dialect.MakeSqlInsert(repoType.tableName, repoType.entity.GetColumns(), data)

	sqlStmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer sqlStmt.Close()
	if dialect.Name() == "mssql" || dialect.Name() == "postgres" {
		var insertedID int64
		err = sqlStmt.QueryRow(args...).Scan(&insertedID)
		if err == nil {
			err = fetchAfterInsertForQueryRow(repoType.entity, dataValue, insertedID)
			if err != nil {
				return err
			}
		}
		return err
	}
	sqlResult, err := sqlStmt.Exec(args...)
	if err != nil {
		m, err1 := migrator.NewMigrator(tx.Db)
		if err1 != nil {
			return err
		}
		shema, err1 := m.GetLoader().LoadFullSchema(tx.Db)
		if err1 != nil {
			return err
		}
		errParse := dialect.ParseError(shema, err)
		if errParse, ok := errParse.(*errors.DbError); ok {
			if errParse.ConstraintName != "" && errParse.ErrorType == errors.ERR_DUPLICATE {
				uk := tenantDB.FindUKConstraint(errParse.ConstraintName)
				if uk != nil {
					errParse.StructName = repoType.entity.GetType().String()
					errParse.Table = repoType.tableName
					errParse.Fields = uk.Fields
					errParse.DbCols = uk.DbCols
					return errParse
				}
			}
			if errParse.ConstraintName != "" && errParse.ErrorType == errors.ERR_REFERENCES {
				fk := migrate.ForeignKeyRegistry.FindByConstraintName(errParse.ConstraintName)
				if fk != nil {

					errParse.Table = fk.ToTable
					errParse.Fields = fk.ToFiels
					errParse.DbCols = fk.ToCols
					errParse.RefTable = fk.FromTable
					errParse.RefCols = fk.FromCols
					errParse.RefFields = fk.FromFiels
					errParse.RefStructName = fk.FromStructName

					return errParse
				}
			}
			errParse.Table = repoType.tableName
			errParse.StructName = repoType.entity.GetType().String()
			errParse.Fields = []string{repoType.entity.GetFieldByColumnName(errParse.DbCols[0])}
		}
		return errParse
	}

	err = r.fetchAfterInsert(sqlResult, repoType.entity, dataValue)
	if err != nil {
		if dbError, ok := err.(*errors.DbError); ok {
			dbError.Table = repoType.tableName
			dbError.StructName = repoType.entity.GetType().String()
			dbError.Fields = []string{repoType.entity.GetFieldByColumnName(dbError.DbCols[0])}
		}
		return err
	}
	return nil
}
func (r *inserter) Insert(db *tenantDB.TenantDB, data interface{}) error {
	dialect := factory.DialectFactory.Create(db.GetDriverName())
	dataValue := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		dataValue = dataValue.Elem()
	}
	repoType, err := r.getEntityInfo(typ)
	if err != nil {
		return err
	}

	sqlText, args := dialect.MakeSqlInsert(repoType.tableName, repoType.entity.GetColumns(), data)
	sqlStmt, err := db.Prepare(sqlText)
	if err != nil {
		return err
	}
	defer sqlStmt.Close()

	switch dialect.Name() {
	case "mssql", "postgres":
		var insertedID int64
		err = sqlStmt.QueryRow(args...).Scan(&insertedID)
		if err == nil {
			err = fetchAfterInsertForQueryRow(repoType.entity, dataValue, insertedID)
		}

	default:
		var res sql.Result
		res, err = sqlStmt.Exec(args...) // ✅ chỗ này đã hợp lệ
		if err == nil {
			err = r.fetchAfterInsert(res, repoType.entity, dataValue)
		}
	}

	if err != nil {
		m, err1 := migrate.NewMigrator(db)
		if err1 != nil {
			return err
		}
		schema, err1 := m.GetLoader().LoadFullSchema(db)
		if err1 != nil {
			return err
		}

		return r.ParserDbError(schema, dialect, err, repoType)
	}

	return nil
}

func (r *inserter) ParserDbError(schema *migrate.DbSchema, dialect dialectCommon.Dialect, err error, repoType *entityInfo) error {

	errParse := dialect.ParseError(schema, err)
	if derr, ok := errParse.(*errors.DbError); ok {
		if derr.ConstraintName != "" {
			if uk := migrate.FindUKConstraint(derr.ConstraintName); uk != nil {
				derr.Table = repoType.tableName
				derr.StructName = repoType.entity.GetType().String()
				derr.Fields = uk.Fields
				derr.DbCols = uk.DbCols
				return derr
			}
		}
		derr.Table = repoType.tableName
		derr.StructName = repoType.entity.GetType().String()
		derr.Fields = []string{repoType.entity.GetFieldByColumnName(derr.DbCols[0])}
		return derr
	}
	return errParse
}

var inserterObj = &inserter{}

func assertSinglePointerToStruct(obj interface{}) error {
	v := reflect.ValueOf(obj)
	t := v.Type()

	depth := 0
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		depth++
	}

	if depth != 1 {
		return fmt.Errorf("Insert: expected pointer to struct (*T), got %d-level pointer", depth)
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("Insert: expected pointer to struct, got pointer to %s", t.Kind())
	}

	return nil
}
func Insert(db *tenantDB.TenantDB, data interface{}) error {
	err := assertSinglePointerToStruct(data)
	if err != nil {
		return err
	}
	m, err := migrate.NewMigrator(db)
	if err != nil {
		return err
	}
	err = m.DoMigrates()
	if err != nil {
		return err
	}

	return inserterObj.Insert(db, data)
}
func InsertWithTx(tx *tenantDB.TenantTx, data interface{}) error {
	err := assertSinglePointerToStruct(data)
	if err != nil {
		return err
	}
	m, err := migrate.NewMigrator(tx.Db)
	if err != nil {
		return err
	}
	err = m.DoMigrates()
	if err != nil {
		return err
	}

	return inserterObj.InsertWithTx(tx, data)
}

type encoderFunc func(v reflect.Value, args *[]interface{})

// func InsertBatchOld[T any](db *tenantDB.TenantDB, data []T) (int64, error) {
// 	if len(data) == 0 {
// 		return 0, nil
// 	}

// 	const maxBatchSize = 200 // nên chọn nhỏ để tránh lỗi "too many placeholders"

// 	m, err := NewMigrator(db)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if err = m.DoMigrates(); err != nil {
// 		return 0, err
// 	}

// 	dialect := dialectFactory.Create(db.GetDriverName())
// 	repoType := inserterObj.getEntityInfo(reflect.TypeOf(data[0]))
// 	tableName := repoType.tableName
// 	columns := []string{}
// 	colDefs := []migrate.ColumnDef{}

// 	// Chỉ lấy các cột không phải Auto
// 	for _, col := range repoType.entity.GetColumns() {
// 		if !col.IsAuto {
// 			columns = append(columns, dialect.Quote(col.Name))
// 			colDefs = append(colDefs, col)
// 		}
// 	}

// 	placeholdersPerRow := "(" + strings.Repeat("?, ", len(colDefs)-1) + "?" + ")"
// 	var totalRows int64

// 	for i := 0; i < len(data); i += maxBatchSize {
// 		end := i + maxBatchSize
// 		if end > len(data) {
// 			end = len(data)
// 		}

// 		batch := data[i:end]
// 		placeholderList := []string{}
// 		args := []interface{}{}

// 		for _, row := range batch {
// 			val := reflect.ValueOf(row)
// 			if val.Kind() == reflect.Ptr {
// 				val = val.Elem()
// 			}
// 			placeholderList = append(placeholderList, placeholdersPerRow)

// 			for _, col := range colDefs {
// 				fieldVal := val.FieldByName(col.Field.Name)
// 				if fieldVal.CanInterface() {
// 					args = append(args, fieldVal.Interface())
// 				} else {
// 					args = append(args, nil)
// 				}
// 			}
// 		}

// 		sql := "INSERT INTO " + dialect.Quote(tableName) + " (" +
// 			strings.Join(columns, ", ") + ") VALUES " +
// 			strings.Join(placeholderList, ", ")

// 		sqlResult, err := db.Exec(sql, args...)
// 		if err != nil {
// 			errParse := dialect.ParseError(err)
// 			if derr, ok := errParse.(*errors.DbError); ok {
// 				derr.Table = tableName
// 				derr.StructName = repoType.entity.GetType().String()
// 				if len(derr.DbCols) > 0 {
// 					derr.Fields = []string{repoType.entity.GetFieldByColumnName(derr.DbCols[0])}
// 				}
// 			}
// 			return totalRows, errParse
// 		}

// 		affected, err := sqlResult.RowsAffected()
// 		if err != nil {
// 			return totalRows, err
// 		}
// 		totalRows += affected
// 	}

// 	return totalRows, nil
// }

var encoderCache sync.Map // map[reflect.Type]func(reflect.Value, *[]interface{})

func getEncoder(t reflect.Type, cols []migrate.ColumnDef) func(reflect.Value, *[]interface{}) {
	if fn, ok := encoderCache.Load(t); ok {
		return fn.(func(reflect.Value, *[]interface{})) // ✅ đúng cách
	}

	fn := func(v reflect.Value, args *[]interface{}) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for _, col := range cols {
			field := v.FieldByName(col.Field.Name)
			if field.CanInterface() {
				*args = append(*args, field.Interface())
			} else {
				*args = append(*args, nil)
			}
		}
	}
	encoderCache.Store(t, fn)
	return fn
}

func InsertBatch[T any](db *tenantDB.TenantDB, data []T) (int64, error) {
	if len(data) == 0 {
		return 0, nil
	}

	const batchSize = 200

	m, err := migrate.NewMigrator(db)
	if err != nil {
		return 0, err
	}
	if err := m.DoMigrates(); err != nil {
		return 0, err
	}

	dialect := factory.DialectFactory.Create(db.GetDriverName())
	repoType, err := inserterObj.getEntityInfo(reflect.TypeOf(data[0]))
	if err != nil {
		return 0, err
	}
	tableName := dialect.Quote(repoType.tableName)

	columns := []string{}
	colDefs := []migrate.ColumnDef{}
	for _, col := range repoType.entity.GetColumns() {
		if !col.IsAuto {
			columns = append(columns, dialect.Quote(col.Name))
			colDefs = append(colDefs, col)
		}
	}
	placeholder := "(" + strings.Repeat("?, ", len(colDefs)-1) + "?" + ")"
	encoder := getEncoder(reflect.TypeOf(data[0]), colDefs)

	var totalRows int64

	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[i:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO ")
		sb.WriteString(tableName)
		sb.WriteString(" (")
		sb.WriteString(strings.Join(columns, ", "))
		sb.WriteString(") VALUES ")

		args := make([]interface{}, 0, len(batch)*len(colDefs))
		for j, row := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(placeholder)
			val := reflect.ValueOf(row)
			encoder(val, &args)
		}

		sql := sb.String()
		sqlResult, err := db.Exec(sql, args...)
		if err != nil {
			shema, err1 := m.GetLoader().LoadFullSchema(db)
			if err1 != nil {
				return totalRows, err
			}

			errParse := dialect.ParseError(shema, err)
			if derr, ok := errParse.(*errors.DbError); ok {
				derr.Table = repoType.tableName
				derr.StructName = repoType.entity.GetType().String()
				if len(derr.DbCols) > 0 {
					derr.Fields = []string{repoType.entity.GetFieldByColumnName(derr.DbCols[0])}
				}
			}
			return totalRows, errParse
		}

		rowsAff, err := sqlResult.RowsAffected()
		if err != nil {
			return totalRows, err
		}
		totalRows += rowsAff
	}

	return totalRows, nil
}
func init() {
	tenantDB.OnDbInsertFunc = func(db *tenantDB.TenantDB, data interface{}) error {
		return Insert(db, data)
	}
	tenantDB.OnTxDbInsertFunc = func(tx *tenantDB.TenantTx, data interface{}) error {
		return InsertWithTx(tx, data)
	}

}
