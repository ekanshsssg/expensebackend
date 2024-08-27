package models

import (
	"gorm.io/gorm"
	"time"
)

type Expense struct {
	ExpenseId          int     `json:"expense_id" gorm:"primaryKey;autoIncrement;not null"`
	GroupId            int     `json:"group_id" gorm:"not null"`
	Category           string  `json:"category" gorm:"not null"`
	Amount             float64 `json:"amount" gorm:"not null"`
	ExpenseDescription string  `json:"description" gorm:"not null"`
	PaidBy             int     `json:"paid_by" gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt

	User User `gorm:"foreignKey:PaidBy"`
}
