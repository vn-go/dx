package models

type User struct {
	ID     int     `db:"pk;auto"`
	UserId *string `db:"size:36;unique"`

	Email *string `db:"uk:uq_email;size:150"`

	Phone *string `db:"size:20"`

	Username     string  `db:"size:250;unique"`
	HashPassword *string `db:"size:100"`
	BaseModel
}

func (u *User) FTS() string {
	return "email,Username,description"
}
