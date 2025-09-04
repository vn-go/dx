package types

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/migate/loader/types"
)

type IMigrator interface {
	GetLoader() types.IMigratorLoader
	Quote(names ...string) string
	GetSqlInstallDb() ([]string, error)
	GetColumnDataTypeMapping() map[reflect.Type]string
	GetGetDefaultValueByFromDbTag() map[string]string
	GetSqlCreateTable(entityType reflect.Type) (string, error)
	GetSqlAddColumn(entityType reflect.Type) (string, error)
	GetSqlAddIndex(entityType reflect.Type) (string, error)
	GetSqlAddUniqueIndex(entityType reflect.Type) (string, error)
	//GetSqlMigrate(entityType reflect.Type) ([]string, error)
	GetSqlAddForeignKey() ([]string, error)
	GetFullScript() ([]string, error)
	//DoMigrate(entityType reflect.Type) error
	DoMigrates() error
}

type migratorInit struct {
	once sync.Once
	val  IMigrator
	err  error
}

type initNewMigrator struct {
	once sync.Once
	val  IMigrator
	err  error
}

var cacheNewMigrator sync.Map

/*
this is Option for AddForeignKey
*/
type CascadeOption struct {
	OnDelete bool
	OnUpdate bool
}
type ForeignKeyInfo struct {
	FromTable      string
	FromCols       []string
	ToTable        string
	ToCols         []string
	FromFiels      []string
	ToFiels        []string
	FromStructName string
	ToStructName   string
	Cascade        CascadeOption
}
type foreignKeyRegistry struct {
	FKMap map[string]*ForeignKeyInfo
}

func (r *foreignKeyRegistry) Register(fk *ForeignKeyInfo) {
	key := fmt.Sprintf("FK_%s__%s_____%s__%s", fk.FromTable, strings.Join(fk.FromCols, "_____"), fk.ToTable, strings.Join(fk.ToCols, "_____"))
	r.FKMap[key] = fk

}
func (r *foreignKeyRegistry) FindByConstraintName(name string) *ForeignKeyInfo {
	if ret, ok := r.FKMap[name]; ok {
		return ret
	}
	return nil
}

var ForeignKeyRegistry = foreignKeyRegistry{
	FKMap: map[string]*ForeignKeyInfo{},
}
