package compiler

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) getSqlInfoFromUnion(node sqlparser.SQLNode) (*types.SqlInfo, error) {
	if stmUnion, ok := node.(*sqlparser.Union); ok {

		ret, err := cmp.getSqlInfoFromUnion(stmUnion.Right)
		if err != nil {
			return nil, err
		}
		previous, err := cmp.getSqlInfoFromUnion(stmUnion.Left)
		if err != nil {
			return nil, err
		}
		ret.UnionPrevious = previous
		ret.UnionType = stmUnion.Type
		ret.Args = internal.UnionCompilerArgs(previous.Args, ret.Args)
		return ret, nil

	}
	if stmSelect, ok := node.(*sqlparser.Select); ok {
		compiler, err := newCompilerFromSqlNode(stmSelect, cmp.dialect)
		if err != nil {
			return nil, err
		}
		ret, err := compiler.getSqlInfo()
		cmp.returnField = compiler.returnField
		cmp.dict = compiler.dict
		if err != nil {
			return nil, err
		}
		cmp.args = internal.UnionCompilerArgs(compiler.args, cmp.args)
		return ret, nil
		//args = compiler.args
	}
	panic(fmt.Sprintf("getSqlInfoFromUnion: not implement %T, see %s", node, `compiler\compiler.getSqlInfoFromUnion.go`))

}
