package controller

import (
	"expensebackend/pkg/config"
	"sort"
	"expensebackend/pkg/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)


func GetActivity(c *gin.Context) {
	_userId, err := strconv.Atoi(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var groups []models.Group
	if err := config.GetDB().Where("group_members=?", _userId).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var activities []models.Activity
	var response []map[string]interface{}

	for _, v := range groups {
		var groupActivities []models.Activity
		if err := config.GetDB().Where("group_id=?", v.GroupId).Find(&groupActivities).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		activities = append(activities, groupActivities...)
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	for _, activity := range activities {

		response = append(response, map[string]interface{}{
			"created_at":  activity.CreatedAt.Format("02 Jan"),
			"description": activity.ActivityDescription,
		})
	}

	if len(response) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"activities": response})
}