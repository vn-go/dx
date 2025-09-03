package common

import (
	"fmt"
	"strings"
)

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
