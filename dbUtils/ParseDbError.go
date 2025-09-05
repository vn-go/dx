package dbutils

import (
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/errors"
	loaderTypes "github.com/vn-go/dx/migate/loader/types"
	"github.com/vn-go/dx/model"
)

func (r *inserter) parseDbError(schema *loaderTypes.DbSchema, dialect types.Dialect, err error, repoType *entity.Entity) error {

	errParse := dialect.ParseError(schema, err)
	if derr, ok := errParse.(*errors.DbErr); ok {

		if derr.ConstraintName != "" {
			if uk := model.ModelRegister.FindUKConstraint(derr.ConstraintName); uk != nil {
				derr.Table = repoType.TableName
				derr.StructName = repoType.EntityType.String()
				derr.Fields = uk.Fields
				derr.DbCols = uk.DbCols
				return derr
			}
		}
		derr.Table = repoType.TableName
		derr.StructName = repoType.EntityType.String()
		derr.Fields = []string{repoType.GetFieldByColumnName(derr.DbCols[0])}
		return derr
	}
	return errParse
}
