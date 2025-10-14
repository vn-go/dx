package postgres

func (m *MigratorPostgres) GetSqlInstallDb(shema string) ([]string, error) {
	return []string{
		/*
			-- Enable GIN index support via pg_trgm (for full text search)
		*/
		"CREATE EXTENSION IF NOT EXISTS pg_trgm;",

		/*-- Enable case-insensitive text (citext)*/
		"CREATE EXTENSION IF NOT EXISTS citext;",
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";",
	}, nil

}
