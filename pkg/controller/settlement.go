package controller

import (
	"expensebackend/pkg/config"
	"expensebackend/pkg/models"
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
)

type SettleInput struct {
	GroupId      int     `json:"group_id" binding:"required"`
	UserPaid     int     `json:"user_paid" binding:"required"`
	UserReceived int     `json:"user_received" binding:"required"`
	Amount       float64 `json:"amount" binding:"required"`
}

func AddSettlement(c *gin.Context) {

	var input SettleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Print(input)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transaction := config.GetDB().Begin()

	result := transaction.Model(&models.Debt{}).Where("group_id=? AND user_who_owes=? AND user_who_owns=?", input.GroupId, input.UserPaid, input.UserReceived).Update("amount", 0)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	result = transaction.Model(&models.Debt{}).Where("group_id=? AND user_who_owes=? AND user_who_owns=?", input.GroupId, input.UserReceived, input.UserPaid).Update("amount", 0)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	settle := &models.Settlement{GroupId: input.GroupId, UserPaid: input.UserPaid, UserReceived: input.UserReceived, Amount: input.Amount}
	result = transaction.Create(settle)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	var userPaidName string
	var userRecivedName string
	var groupName string
	result = transaction.Model(&models.User{}).Where("id=?", input.UserPaid).Select("name").Scan(&userPaidName)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	result = transaction.Model(&models.User{}).Where("id=?", input.UserReceived).Select("name").Scan(&userRecivedName)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	result = transaction.Model(&models.Group{}).Where("group_id=?", input.GroupId).Select("name").Scan(&groupName)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	activityDesc := fmt.Sprintf("%s paid %s amount ₹ %s in '%s", userPaidName, userRecivedName, strconv.FormatFloat(input.Amount, 'f', 2, 64), groupName)
	activity := &models.Activity{GroupId: input.GroupId, ActivityDescription: activityDesc}
	if err := transaction.Create(activity).Error; err != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	transaction.Commit()

	var group models.Group
	// var groupName string

	if err := config.GetDB().Where("group_id=?", input.GroupId).First(&group).Error; err != nil {
		groupName = ""
	} else {
		groupName = group.Name
	}

	var users []models.User
	var members []string
	var userPaid models.User
	var userRecived models.User
	if err := config.GetDB().Where("id=?", input.UserPaid).First(&userPaid).Error; err == nil {

		users = append(users, userPaid)
		members = append(members, userPaid.EmailId)
		if err = config.GetDB().Where("id=?", input.UserReceived).First(&userRecived).Error; err == nil {
			users = append(users, userRecived)
			go func() {
				members = append(members, userRecived.EmailId)
				message := fmt.Sprintf("%s paid %s amount ₹ %s in %s.", users[0].Name, users[1].Name, strconv.FormatFloat(input.Amount, 'f', 2, 64), groupName)
				Maill(members, "Settlement done", message)
			}()

		} else {
			fmt.Println("Error fetching user")
		}

	} else {
		fmt.Println("Error fetching user")
	}

	c.JSON(http.StatusOK, gin.H{"success": "Settlement Done"})

}