package models

type Salary struct {
	BaseModel
	ID        int     `db:"pk;auto"`
	UserID    int     `db:"idx:idx_salary_user"`
	Month     string  `db:"type:char(7);idx:idx_salary_month"` // e.g. 2024-07
	Base      float64 `db:"type:decimal(15,2)"`
	Bonus     float64 `db:"type:decimal(15,2)"`
	Deduction float64 `db:"type:decimal(15,2)"`
}
