package dx

import (
	common "github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

var Open = tenantDB.Open
var ModelRegistry = common.ModelRegistry
