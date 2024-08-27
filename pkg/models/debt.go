package models

type Debt struct {
	Id          int     `json:"id" gorm:"not null;primaryKey;autoIncrement"`
	GroupId     int     `json:"group_id" gorm:"not null"`
	UserWhoOwns int     `json:"user_who_owns" gorm:"not null"`
	UserWhoOwes  int     `json:"user_who_ows" gorm:"not null"`
	Amount      float64 `json:"amount" gorm:"not null"`

	// User    User    `gorm:"foreignKey:UserWhoOwns,UserWhoOws"`
}
