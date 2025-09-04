package models

type Attendance struct {
	BaseModel
	ID       int    `db:"pk;auto"`
	UserID   int    `db:"idx:idx_att_user"`
	Date     string `db:"type:date;idx:idx_att_date"`
	CheckIn  string `db:"type:time"`
	CheckOut string `db:"type:time"`
}
