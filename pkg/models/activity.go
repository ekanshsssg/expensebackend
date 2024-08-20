package models

import (
	_ "github.com/jinzhu/gorm"
	"gorm.io/gorm"
)

type Activity struct {
	gorm.Model
	GroupId     int `json:"group_id" gorm:"not null"`
	ActivityDescription  string `json:"activity_description" gorm:"unique;not null"`
}