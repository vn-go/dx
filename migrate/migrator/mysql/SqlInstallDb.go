package mysql

func (m *MigratorMySql) GetSqlInstallDb(schema string) ([]string, error) {
	return []string{}, nil // no thing to do for MySQL
}
