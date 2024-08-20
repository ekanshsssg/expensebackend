package controller

import (
	"database/sql"
	"expensebackend/pkg/config"

	"expensebackend/pkg/models"
	"fmt"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/jinzhu/gorm"
	// "gorm.io/gorm"
)

type CreateGroupInput struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description" binding:"required"`
	Created_by    int    `json:"created_by" binding:"required"`
	Group_members int    `json:"group_members" binding:"required"`
}

type MembersInput struct {
	GroupId int   `json:"group_id" binding:"required"`
	UserId  int   `json:"user_id" binding:"required"`
	Members []int `json:"members" binding:"required"`
}

func CreateGroup(c *gin.Context) {
	var input CreateGroupInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction := config.GetDB().Begin()

	var maxGroupID sql.NullInt64
	if err := transaction.Model(&models.Group{}).Select("MAX(group_id)").Scan(&maxGroupID).Error; err != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var newGroupId int
	if !maxGroupID.Valid {
		newGroupId = 1
	} else {
		newGroupId = int(maxGroupID.Int64) + 1
	}

	group := models.Group{GroupId: newGroupId, Name: input.Name, Description: input.Description, CreatedBy: input.Created_by, GroupMembers: input.Group_members}

	result := transaction.Create(&group)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	var createByName string

	if err := transaction.Model(&models.User{}).Where("id=?", input.Created_by).Select("name").Scan(&createByName).Error; err != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activityDesc := fmt.Sprintf("%s created a new group '%s' ", createByName, group.Name)
	activity := &models.Activity{GroupId: newGroupId, ActivityDescription: activityDesc}
	if err := transaction.Create(activity).Error; err != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	transaction.Commit()

	c.JSON(http.StatusOK, gin.H{
		"groupid": newGroupId,
		"message": "Group created successgully",
	})
}

func GetGroups(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.Atoi(userIdStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var groups []models.Group
	if err := config.GetDB().Where("group_members=?", userId).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query execution failed"})
		return
	}

	c.JSON(http.StatusOK, groups)
}










