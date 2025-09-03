package common

import (
	"reflect"
	"sync"
)

type IMigrator interface {
	GetLoader() IMigratorLoader
	Quote(names ...string) string
	GetSqlInstallDb() ([]string, error)
	GetColumnDataTypeMapping() map[reflect.Type]string
	GetGetDefaultValueByFromDbTag() map[string]string
	GetSqlCreateTable(entityType reflect.Type) (string, error)
	GetSqlAddColumn(entityType reflect.Type) (string, error)
	GetSqlAddIndex(entityType reflect.Type) (string, error)
	GetSqlAddUniqueIndex(entityType reflect.Type) (string, error)
	GetSqlMigrate(entityType reflect.Type) ([]string, error)
	GetSqlAddForeignKey() ([]string, error)
	GetFullScript() ([]string, error)
	DoMigrate(entityType reflect.Type) error
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
