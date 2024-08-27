package models

import (
	"gorm.io/gorm"
)

type Settlement struct {
	gorm.Model
	GroupId      int     `json:"group_id" gorm:"not null"`
	UserPaid     int     `json:"user_paid" gorm:"unique;not null"`
	UserReceived int     `json:"user_received" gorm:"not null"`
	Amount       float64 `json:"amount" gorm:"not null"`
}
