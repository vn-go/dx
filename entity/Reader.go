package entity

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/vn-go/dx/internal"
)

type readerType struct {
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
func (u *readerType) ParseStruct(t reflect.Type, parentIndexOfField []int) ([]ColumnDef, error) {

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var cols []ColumnDef
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
this function will parse the db tag of a field
if tag not found, it will return ColumnDef with default values
look at the example below:

	ColumnDef {
		Name:     name of the field, // the other info is default
	}
*/
func (u *readerType) ParseTagFromStruct(field reflect.StructField, parentIndexOfField []int) ColumnDef {
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

	col := ColumnDef{
		Name:         internal.Utils.SnakeCase(field.Name),
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
func (u *readerType) parseLengthFromType(typeStr string) *int {
	re := regexp.MustCompile(`\((\d+)\)`)
	match := re.FindStringSubmatch(typeStr)
	if len(match) == 2 {
		if length, err := strconv.Atoi(match[1]); err == nil {
			return &length
		}
	}
	return nil
}

var Reader = &readerType{}
