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
	// var ret *types.SqlInfo
	// var err error
	//var args internal.CompilerArgs

	// if left, ok := stmUnion.Left.(*sqlparser.Select); ok {
	// 	compiler, err := newCompilerFromSqlNode(left, cmp.dialect)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	ret, err = compiler.getSqlInfo()
	// 	cmp.returnField = compiler.returnField
	// 	cmp.dict = compiler.dict
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	//args = compiler.args
	// } else if left, ok := stmUnion.Left.(*sqlparser.Union); ok {
	// 	ret, err = cmp.getSqlInfoFromUnion(left)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// } else {
	// 	panic(fmt.Sprintf("compiler.getSqlInfo: not support %T", stmUnion.Left))
	// }

	// if right, ok := stmUnion.Right.(*sqlparser.Select); ok {
	// 	// var next *types.SqlInfo
	// 	// cmp.args = internal.FillArrayToEmptyFields[internal.CompilerArgs, internal.SqlArgs](internal.CompilerArgs{})
	// 	// next, err := cmp.getSqlInfoBySelect(right)
	// 	compiler, err := newCompilerFromSqlNode(right, cmp.dialect)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	next, err := compiler.getSqlInfo()

	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if err != nil {
	// 		return nil, err
	// 	} else {
	// 		if ret.UnionLast == nil {
	// 			ret.UnionType = stmUnion.Type
	// 			ret.UnionNext = next
	// 			//ret.Args = internal.UnionCompilerArgs(ret.Args, next.Args)
	// 		} else {
	// 			last := ret.UnionNext
	// 			for ret.UnionNext.UnionNext != nil {
	// 				last = last.UnionNext
	// 			}
	// 			last.UnionType = stmUnion.Type
	// 			last.UnionNext = next
	// 		}
	// 		// if ret.UnionNext != nil {
	// 		// 	ret.UnionNext.UnionType = stmUnion.Type
	// 		// 	ret.UnionNext.UnionNext = next
	// 		// 	//ret.Args = internal.UnionCompilerArgs(ret.Args, next.Args)
	// 		// } else {
	// 		// 	ret.UnionType = stmUnion.Type
	// 		// 	ret.UnionNext = next
	// 		// 	//ret.Args = internal.UnionCompilerArgs(ret.Args, next.Args)
	// 		// }

	// 	}
	// } else {
	// 	panic(fmt.Sprintf("compiler.getSqlInfo: not support %T", stmUnion.Left))
	// }
	// //fmt.Println(mainArgs)
	// //ret.Args = args
	// return ret, nil
}
