package models

type Position struct {
	Code  string `db:"size:100;uk:uq_pos_code"`
	Name  string `db:"size:100;uk:uq_pos_name"`
	ID    int    `db:"pk;auto"`
	Title string `db:"size:100;uk:uq_pos_title"`
	Level int

	BaseModel
}
