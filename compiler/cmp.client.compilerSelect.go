package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *cmpSelectorType) resolevSelector(dialect types.Dialect, outputFields map[string]string, n sqlparser.SelectExprs) (string, error) {
	// if _, ok := n.(sqlparser.StarExpr); ok {
	// 	return "", NewCompilerError(fmt.Sprintf("'%s' is invalid expession"))
	// }
	strFields := []string{}
	for _, x := range n {
		f, err := cmp.resolve(dialect, outputFields, x)
		if err != nil {
			return "", err
		}
		strFields = append(strFields, f)

	}
	return strings.Join(strFields, ","), nil
}
func (cmp *cmpSelectorType) resolve(dialect types.Dialect, outputFields map[string]string, n sqlparser.SQLNode) (string, error) {

	if isDebugMode {
		panic(fmt.Sprintf("Not implement %T, see 'resolve' in %s", n, `compiler\cmp.client.compilerSelect.go`))
	} else {
		return "", NewCompilerError(fmt.Sprintf("'%s' is invalid syntax"))
	}

}
