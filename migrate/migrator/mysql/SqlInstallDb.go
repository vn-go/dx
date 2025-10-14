package mysql

func (m *MigratorMySql) GetSqlInstallDb() ([]string, error) {
	return []string{}, nil // no thing to do for MySQL
}
