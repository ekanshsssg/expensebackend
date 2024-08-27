package models

type Expensemember struct {
	Id            int     `json:"id" gorm:"not null:primaryKey:autoIncrement"`
	ExpenseId     int     `json:"expense_id" gorm:"not null"`
	GroupId       int     `json:"group_id" gorm:"not null"`
	ExpenseMember int     `json:"expense_member" gorm:"not null"`
	Amount        float64 `json:"amount" gorm:"not null"`

	User    User    `gorm:"foreignKey:ExpenseMember"`
	Expense Expense `gorm:"foreignKey:ExpenseId"`
}
