package dbutils

import (
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/errors"
	loaderTypes "github.com/vn-go/dx/migrate/loader/types"
)

func (r *inserter) parseDbError(schema *loaderTypes.DbSchema, dialect types.Dialect, err error, repoType *entity.Entity) error {

	errParse := dialect.ParseError(schema, err)
	if derr, ok := errParse.(*errors.DbErr); ok {

		if derr.ConstraintName != "" {
			if uk := types.FindUKConstraint(derr.ConstraintName); uk != nil {
				derr.Table = repoType.TableName
				derr.StructName = repoType.EntityType.String()
				derr.Fields = uk.Fields
				derr.DbCols = uk.DbCols
				return derr
			}
			if info, ok := schema.PrimaryKeys[derr.ConstraintName]; ok {
				return r.parsePrimaryKeysError(derr, info, repoType)

			}
		}
		derr.Table = repoType.TableName
		derr.StructName = repoType.EntityType.String()
		if len(derr.DbCols) > 0 {
			derr.Fields = []string{repoType.GetFieldByColumnName(derr.DbCols[0])}
		}

		return derr
	}
	return errParse
}

func (r *inserter) parsePrimaryKeysError(derr *errors.DbErr, info loaderTypes.ColumnsInfo, repoType *entity.Entity) error {
	derr.Table = repoType.TableName
	derr.StructName = repoType.EntityType.String()
	derr.Fields = []string{}
	derr.DbCols = []string{}
	for _, x := range info.Columns {

		derr.DbCols = append(derr.DbCols, x.Name)
		for _, y := range repoType.Cols {
			if y.Name == x.Name {
				derr.Fields = append(derr.Fields, y.Field.Name)
				break
			}
		}
	}

	return derr
}
