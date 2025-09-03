package dx
import "github.com/vn-go/dx/internal"
func SetManagerDb(driver string, dbName string) error {
	return internal.TenantDbManagerInstance.SetManagerDb(driver, dbName)
}
