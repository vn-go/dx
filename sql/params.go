package sql

import (
	"fmt"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type param struct {
}

func (p *param) sqlVal(expr *sqlparser.SQLVal, injector *injector) (*compilerResult, error) {
	switch expr.Type {
	case sqlparser.IntVal:
		val, err := internal.Helper.ToIntFromBytes(expr.Val)
		if err != nil {
			return nil, err // invalid argument for dynamic param function
		}
		newArgs := argument{
			val:        val,
			isConstant: true,
		}
		injector.args = append(injector.args, newArgs)
		newArgs.index = len(injector.args)
		return &compilerResult{
			OriginalContent: string(expr.Val),
			Content:         injector.dialect.ToParam(len(injector.args), sqlparser.IntVal),
			Args: arguments{
				newArgs,
			},
			IsExpression: true,
		}, nil
	case sqlparser.BitVal:
		val := internal.Helper.ToBoolFromBytes(expr.Val)
		newArg := argument{
			val:        val,
			isConstant: true,
		}
		injector.args = append(injector.args, newArg)
		newArg.index = len(injector.args)
		return &compilerResult{
			OriginalContent: string(expr.Val),
			Content:         injector.dialect.ToParam(len(injector.args), sqlparser.IntVal),
			Args: arguments{
				newArg,
			},
			IsExpression: true,
		}, nil
	case sqlparser.FloatVal:
		val := internal.Helper.ToBoolFromBytes(expr.Val)
		newArg := argument{
			val:        val,
			isConstant: true,
		}
		injector.args = append(injector.args, newArg)
		newArg.index = len(injector.args)
		return &compilerResult{
			OriginalContent: string(expr.Val),
			Content:         injector.dialect.ToParam(len(injector.args), sqlparser.IntVal),
			Args:            arguments{newArg},
			IsExpression:    true,
		}, nil
	case sqlparser.StrVal:
		val := string(expr.Val)
		newArg := argument{
			val:        val,
			isConstant: true,
		}
		injector.args = append(injector.args, newArg)
		newArg.index = len(injector.args)
		return &compilerResult{
			OriginalContent: val,
			Content:         injector.dialect.ToParam(len(injector.args), sqlparser.IntVal),
			Args:            arguments{newArg},
			IsExpression:    true,
		}, nil
	default:
		panic(fmt.Sprintf("unsupported type: %d. See param.sqlVal, file %s", expr.Type, `sql\params.go`))
	}

}

func (p *param) extract(node sqlparser.SQLNode) *sqlparser.SQLVal {
	switch x := node.(type) {
	case *sqlparser.SQLVal:
		return x
	case *sqlparser.AliasedExpr:
		return p.extract(x.Expr)

	default:
		panic(fmt.Sprintf("unsupported type: %T. See param.extract, file %s ", x, `sql\params.go`))
	}

}
func (p *param) funcExpr(x *sqlparser.FuncExpr, injector *injector) (*compilerResult, error) {
	if x.Name.String() == GET_PARAMS_FUNC { // this is dynamic param sent by calling compiled function

		val := p.extract(x.Exprs[0])
		indexOfArg, err := internal.Helper.ToIntFromBytes(val.Val)
		if err != nil {
			return nil, err // invalid argument for dynamic param function
		}
		newArgs := argument{
			index:      indexOfArg,
			isConstant: false,
		}
		injector.args = append(injector.args, newArgs)
		return &compilerResult{
			OriginalContent: "?",
			Content:         injector.dialect.ToParam(len(injector.args), -1), // -1 means dynamic param type
			IsExpression:    true,
			Args:            arguments{newArgs},
			
		}, nil
	}
	if x.Name.String() == internal.FnMarkSpecialTextArgs { // this is a function call to get current timestamp
		val := p.extract(x.Exprs[0])
		indexOfArg, err := internal.Helper.ToIntFromBytes(val.Val)
		if err != nil {
			return nil, err // invalid argument for dynamic param function
		}
		newArgs := argument{
			val:        injector.textParams[indexOfArg],
			isConstant: true,
		}
		injector.args = append(injector.args, newArgs)
		return &compilerResult{
			OriginalContent: "?",
			Content:         injector.dialect.ToParam(len(injector.args), -1), // -1 means dynamic param type
			IsExpression:    true,
			Args:            arguments{newArgs},
		}, nil
	}
	panic(fmt.Sprintf("unsupported function: %s. see param.funcExpr, file %s", x.Name.String(), `sql\params.go`))
}

var params = &param{}
