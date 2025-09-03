package common

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	pluralizeLib "github.com/gertd/go-pluralize"
	"github.com/vn-go/dx/internal"
)

var pluralize = pluralizeLib.NewClient()

/*
this struct is used when DEV wants to make their struct as Model
example:

	type BaseModel struct {
		CreatedAt time.Time
	}
	type User struct {
		Entity `db:"table:users"` // <- if db tag not defined, table name is converted to SnakeCase of struct name and pluralized
		BaseModel
		ID int `db:"pk"`
	}
*/
type Entity struct {
	EntityType               reflect.Type
	tableName                string
	Cols                     []internal.ColumnDef          //<-- list of all columns
	mapCols                  map[string]internal.ColumnDef //<-- used for faster access to column by name
	PrimaryConstraints       map[string][]internal.ColumnDef
	uniqueConstraints        map[string][]internal.ColumnDef
	indexConstraints         map[string][]internal.ColumnDef
	buildUniqueConstraints   map[string][]internal.ColumnDef
	cacheGetAutoValueColumns sync.Map
	DbTableName              string
}
type initGetAutoValueColumns struct {
	once sync.Once
	val  []internal.ColumnDef
}

func (e *Entity) GetAutoValueColumns() []internal.ColumnDef {
	actual, _ := e.cacheGetAutoValueColumns.LoadOrStore(e.EntityType, &initGetAutoValueColumns{})
	init := actual.(*initGetAutoValueColumns)
	init.once.Do(func() {
		init.val = []internal.ColumnDef{}
		for _, col := range e.Cols {
			if col.IsAuto {
				init.val = append(init.val, col)
			}
		}
	})
	return init.val
}
func (e *Entity) GetType() reflect.Type {
	return e.EntityType
}
func (e *Entity) GetColumns() []internal.ColumnDef {
	return e.Cols
}
func (e *Entity) TableName() string {
	return e.tableName
}
func (e *Entity) GetFieldByColumnName(colName string) string {
	col, ok := e.mapCols[colName]
	if ok {
		return col.Field.Name
	}
	return ""
}

func (e *Entity) GetIndexConstraints() map[string][]internal.ColumnDef {
	if e.indexConstraints == nil || len(e.indexConstraints) == 0 {
		e.indexConstraints = map[string][]internal.ColumnDef{}
		for _, col := range e.Cols {
			if col.IndexName != "" {
				e.indexConstraints[col.IndexName] = append(e.indexConstraints[col.IndexName], col)
			}
		}
	}
	return e.indexConstraints
}
func (e *Entity) GetUniqueConstraints() map[string][]internal.ColumnDef {
	if len(e.uniqueConstraints) == 0 {
		e.uniqueConstraints = map[string][]internal.ColumnDef{}
		for _, col := range e.Cols {
			if col.UniqueName != "" {
				e.uniqueConstraints[col.UniqueName] = append(e.uniqueConstraints[col.UniqueName], col)
			}
		}
	}
	return e.uniqueConstraints
}

type initGetBuildUniqueConstraints struct {
	once sync.Once
	val  map[string][]internal.ColumnDef
}

var cacheGetBuildUniqueConstraints = sync.Map{}

func (e *Entity) GetBuildUniqueConstraints() map[string][]internal.ColumnDef {
	actual, _ := cacheGetBuildUniqueConstraints.LoadOrStore(e.EntityType, &initGetBuildUniqueConstraints{})
	init := actual.(*initGetBuildUniqueConstraints)
	init.once.Do(func() {
		init.val = e.getBuildUniqueConstraints()
	})
	return init.val
}
func (e *Entity) getBuildUniqueConstraints() map[string][]internal.ColumnDef {
	if len(e.buildUniqueConstraints) == 0 {
		e.buildUniqueConstraints = map[string][]internal.ColumnDef{}
		for _, constraint := range e.GetUniqueConstraints() {
			tableName := e.tableName

			cols := []string{}
			for _, col := range constraint {
				cols = append(cols, col.Name)
			}
			key := "UQ_" + tableName + "__" + strings.Join(cols, "___")
			e.buildUniqueConstraints[key] = constraint
		}
	}
	return e.buildUniqueConstraints
}

type utils struct {
	cachePlural      sync.Map //<-- cache for pluralize
	cacheToSnakeCase sync.Map //<-- cache for SnakeCase
}

var dbTagPattern = regexp.MustCompile(`([a-zA-Z]+)(\((.*?)\))?`)

/*
this function will parse the db tag of a field
if tag not found, it will return ColumnDef with default values
look at the example below:

	ColumnDef {
		Name:     name of the field, // the other info is default
	}
*/
func (u *utils) ParseTagFromStruct(field reflect.StructField, parentIndexOfField []int) internal.ColumnDef {
	/*
		Parses struct tag into a ColumnDef following "db" tag convention.
		Supports alternative fallback to "gorm" for compatibility.
		Tag format examples:
			`db:"column:user_name;pk;auto;default:now;type:string(100)"`

		Recognized keys:
			- column         : column name
			- pk, primary    : primary key
			- uk, unique     : unique constraint
			- idx, index     : normal index
			- auto, autoincrement : auto increment flag
			- default        : default value
			- type           : type string with optional length (e.g., string(100))
			- size, length   : explicitly defined length
	*/

	tagStr := field.Tag.Get("db")
	if tagStr == "" {
		tagStr = field.Tag.Get("gorm") // fallback
	}

	col := internal.ColumnDef{
		Name:         u.SnakeCase(field.Name),
		Field:        field,
		Nullable:     field.Type.Kind() == reflect.Ptr,
		IndexOfField: append(parentIndexOfField, field.Index[0]),
	}

	if tagStr == "" {
		return col
	}

	tags := strings.Split(tagStr, ";")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		key := tag
		if key == "index" || key == "idx" {
			col.IndexName = col.Name
			continue
		}
		if key == "uk" || key == "unique" {
			col.UniqueName = col.Name
			continue
		}
		if strings.Contains(key, "index:") {
			col.IndexName = key[6:]
			continue
		}
		if strings.Contains(key, "idx:") {
			col.IndexName = key[4:]
			continue
		}
		if strings.Contains(key, "index(") {
			key = strings.Split(key, "(")[1]
			col.IndexName = strings.Split(key, "(")[0]
			continue
		}
		if strings.Contains(key, "idx(") {
			key = strings.Split(key, "(")[1]
			col.IndexName = strings.Split(key, "(")[0]
			continue
		}

		if strings.Contains(key, "unique:") {
			col.UniqueName = key[7:]
			continue
		}
		if strings.Contains(key, "uk:") {
			col.UniqueName = key[3:]
			continue
		}
		if strings.Contains(key, "unique(") {
			key = strings.Split(key, "(")[1]
			col.UniqueName = strings.Split(key, "(")[0]
			continue
		}
		if strings.Contains(key, "uk(") {
			key = strings.Split(key, "(")[1]
			col.UniqueName = strings.Split(key, "(")[0]
			continue
		}
		if tag == "" {
			continue
		}

		var val string

		if strings.Contains(tag, ":") {
			parts := strings.SplitN(tag, ":", 2)
			key = strings.ToLower(parts[0])
			val = strings.TrimSpace(parts[1])
		} else if strings.Contains(tag, "(") && strings.HasSuffix(tag, ")") {
			// e.g., column(name)
			parts := strings.SplitN(tag, "(", 2)
			key = strings.ToLower(parts[0])
			val = strings.TrimSuffix(parts[1], ")")
		} else {
			key = strings.ToLower(tag)
			val = ""
		}

		switch key {
		case "column":
			if val != "" {
				col.Name = val
			}
		case "pk", "primary", "primarykey":
			col.PKName = "PK"
		case "uk", "unique":
			col.UniqueName = col.Name
		case "index":
			col.IndexName = col.Name

		case "auto", "autoincrement":
			col.IsAuto = true
		case "default":
			col.Default = val
		case "type":
			if length := u.parseLengthFromType(val); length != nil {
				col.Length = length
			}
		case "size", "length":
			if l, err := strconv.Atoi(val); err == nil {
				col.Length = &l
			}
		}
	}

	return col
}

func (u *utils) parseLengthFromType(typeStr string) *int {
	re := regexp.MustCompile(`\((\d+)\)`)
	match := re.FindStringSubmatch(typeStr)
	if len(match) == 2 {
		if length, err := strconv.Atoi(match[1]); err == nil {
			return &length
		}
	}
	return nil
}

/*
This function will parse the struct tag and return a slice of ColumnDef
Example:

	type BaseModel struct {
		CreatedAt time.Time
	}
	type User struct {
		Entity `db:"table:users"` // <- if db tag not defined, table name is converted to SnakeCase of struct name and pluralized
		BaseModel
		ID int `db:"pk"`
	}
*/
func (u *utils) ParseStruct(t reflect.Type, parentIndexOfField []int) ([]internal.ColumnDef, error) {

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var cols []internal.ColumnDef
	for i := 0; i < t.NumField(); i++ {

		f := t.Field(i)
		if f.Type.Kind() == reflect.Func {
			fmt.Println(f.Name)
			continue
		}

		if f.Anonymous {
			if f.Type == reflect.TypeOf(Entity{}) {
				continue // skip embedded Entity
			}
			nextParentFieldIndex := append(parentIndexOfField, f.Index[0])
			subCols, _ := u.ParseStruct(f.Type, nextParentFieldIndex)
			cols = append(cols, subCols...)
			continue
		} else if f.IsExported() {
			cols = append(cols, u.ParseTagFromStruct(f, parentIndexOfField))
		}
	}
	return cols, nil
}

/*
this function will return a map of primary key columns with their names
@return map[string][]internal.ColumnDef //<-- map[Primary key constraint name][]internal.ColumnDef
*/
func (u *utils) GetPrimaryKey(e *Entity) map[string][]internal.ColumnDef {
	m := make(map[string][]internal.ColumnDef)
	for _, col := range e.Cols {
		if col.PKName != "" {
			m[col.PKName] = append(m[col.PKName], col)
		}
	}
	return m
}

/*
this function will return a map of unique key columns with their names
@return map[string][]internal.ColumnDef //<-- map[Unique constraint name][]internal.ColumnDef
*/
func (u *utils) GetUnique(e *Entity) map[string][]internal.ColumnDef {
	m := make(map[string][]internal.ColumnDef)
	for _, col := range e.Cols {
		if col.UniqueName != "" {
			m[col.UniqueName] = append(m[col.UniqueName], col)
		}
	}
	return m
}

/*
this function will return a map of index columns with their names
@return map[string][]internal.ColumnDef //<-- map[Index constraint name][]internal.ColumnDef
*/
func (u *utils) GetIndex(e *Entity) map[string][]internal.ColumnDef {
	m := make(map[string][]internal.ColumnDef)
	for _, col := range e.Cols {
		if col.IndexName != "" {
			m[col.IndexName] = append(m[col.IndexName], col)
		}
	}
	return m
}

/*
this function will return a string of unsynced columns
unsynced columns are columns in Entity that do not exist in the db table
@dbColumnName is a list of column names in the db table
*/
func (u *utils) GetUnSyncColumns(e *Entity, dbColumnName []string) string {
	dbCols := make(map[string]bool)
	for _, c := range dbColumnName {
		dbCols[c] = true
	}

	var unsync []string
	for _, col := range e.Cols {
		if _, found := dbCols[col.Name]; !found {
			unsync = append(unsync, col.Name)
		}
	}
	return strings.Join(unsync, ", ")
}
func (u *utils) Pluralize(txt string) string {
	if v, ok := u.cachePlural.Load(txt); ok {
		return v.(string)
	}
	txt = strings.ToLower(txt)
	ret := pluralize.Plural(txt)
	u.cachePlural.Store(txt, ret)

	return ret
}
func (u *utils) SnakeCase(str string) string {
	if v, ok := u.cacheToSnakeCase.Load(str); ok {
		return v.(string)
	}
	var result []rune
	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) &&
			(unicode.IsLower(rune(str[i-1])) || (i+1 < len(str) && unicode.IsLower(rune(str[i+1])))) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	ret := string(result)
	u.cacheToSnakeCase.Store(str, ret)
	return ret
}

var utilsInstance = &utils{}

const SkipDefaulValue = "vdb::skip"
