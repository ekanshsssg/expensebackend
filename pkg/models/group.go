package models

import (
	"time"
	"gorm.io/gorm"
)

type Group struct {
	GroupId      int    `json:"group_id" gorm:"not null;primaryKey"`
	Name         string `json:"name" gorm:"not null"`
	Description  string `json:"description" gorm:"not null"`
	CreatedBy    int    `json:"created_by" gorm:"not null"`
	GroupMembers int    `json:"group_members" gorm:"not null;primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt 
	User    User    `gorm:"foreignKey:GroupMembers"`
}
