package sql

import (
	"fmt"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type param struct {
}

func (p *param) extract(node sqlparser.SQLNode) *sqlparser.SQLVal {
	switch x := node.(type) {
	case *sqlparser.SQLVal:
		return x
	case *sqlparser.AliasedExpr:
		return p.extract(x.Expr)

	default:
		panic(fmt.Sprintf("unsupported type: %T. See param.extract, file ", x, `sql\params.go`))
	}

}
func (p *param) funcExpr(x *sqlparser.FuncExpr, injector *injector) (*compilerResult, error) {
	if x.Name.String() == GET_PARAMS_FUNC { // this is dynamic param sent by calling compiled function

		val := p.extract(x.Exprs[0])
		indexOfArg, err := internal.Helper.ToIntFormBytes(*&val.Val)
		if err != nil {
			return nil, err // invalid argument for dynamic param function
		}
		injector.args = append(injector.args, argument{
			index: indexOfArg,
		})
		return &compilerResult{
			OriginalContent: "?",
			Content:         injector.dialect.ToParam(len(injector.args), -1), // -1 means dynamic param type
		}, nil
	}
	if x.Name.String() == internal.FnMarkSpecialTextArgs { // this is a function call to get current timestamp
		val := p.extract(x.Exprs[0])
		indexOfArg, err := internal.Helper.ToIntFormBytes(*&val.Val)
		if err != nil {
			return nil, err // invalid argument for dynamic param function
		}
		injector.args = append(injector.args, argument{
			val: injector.textParams[indexOfArg],
		})
		return &compilerResult{
			OriginalContent: "?",
			Content:         injector.dialect.ToParam(len(injector.args), -1), // -1 means dynamic param type
		}, nil
	}
	panic(fmt.Sprintf("unsupported function: %s. see param.funcExpr, file ", x.Name.String(), `sql\params.go`))
}

var params = &param{}
