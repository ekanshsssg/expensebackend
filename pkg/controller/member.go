package controller

import (
	// "database/sql"
	"errors"
	"expensebackend/pkg/config"
	// "sort"

	// "expensebackend/pkg/controller"
	"expensebackend/pkg/models"
	"fmt"

	// "encoding/csv"
	// "fmt"
	// "gorm.io/driver/sql"
	// "fmt"
	"net/http"
	// "strconv"

	// "os"
	// "time"
	"github.com/gin-gonic/gin"

	_ "github.com/jinzhu/gorm"
	"gorm.io/gorm"
)

// type MembersInput struct {
// 	GroupId int   `json:"group_id" binding:"required"`
// 	UserId  int   `json:"user_id" binding:"required"`
// 	Members []int `json:"members" binding:"required"`
// }

func GetMember(c *gin.Context) {
	emailId := c.Query("emailId")
	groupId := c.Query("groupId")
	fmt.Print(emailId)
	var user models.User
	if err := config.GetDB().Where("email_id = ?", emailId).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var group models.Group
	if err := config.GetDB().Unscoped().Where("group_id = ? AND group_members=?", groupId, user.ID).First(&group).Error; err == nil {
		if group.DeletedAt.Valid {
			c.JSON(http.StatusOK, gin.H{
				"name":     user.Name,
				"memberId": user.ID,
				"emailId":  user.EmailId,
			})
			return
		} else {
			c.JSON(http.StatusConflict, gin.H{"error": "user already in your group"})
			return
		}

	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"name":     user.Name,
				"memberId": user.ID,
				"emailId":  user.EmailId,
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})

}

func AddMembers(c *gin.Context) {

	var input MembersInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// transaction := config.GetDB().Begin()
	var group models.Group
	if err := config.GetDB().Where("created_by=? AND group_id=?", input.UserId, input.GroupId).First(&group).Error; err != nil {
		if gorm.ErrRecordNotFound == err {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You can not add members"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var userAddedBy models.User
	var userAddedByName string
	if err := config.GetDB().Where("id=?", input.UserId).First(&userAddedBy).Error; err != nil {
		userAddedByName = ""
	} else {
		userAddedByName = "by " + userAddedBy.Name
	}

	err := config.GetDB().Transaction(func(tx *gorm.DB) error {

		for i := 0; i < len(input.Members); i++ {
			var existing models.Group
			if err := tx.Unscoped().Where("group_id=? AND group_members=?", input.GroupId, input.Members[i]).First(&existing).Error; err == nil {
				if existing.DeletedAt.Valid {
					// Member is soft deleted, set DeletedAt to NULL
					existing.DeletedAt = gorm.DeletedAt{}
					if err := tx.Save(&existing).Error; err != nil {
						return err
					}
				}
			} else {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					addGroupRow := &models.Group{GroupId: group.GroupId, Name: group.Name, Description: group.Description, CreatedBy: input.UserId, GroupMembers: input.Members[i]}
					if err := tx.Create(addGroupRow).Error; err != nil {
						return err
					}
				}
			}

			var memName string

			if err := config.GetDB().Model(&models.User{}).Where("id=?", input.Members[i]).Select("name").Scan(&memName).Error; err != nil {
				return err
			}

			activityDesc := fmt.Sprintf("%s was added in '%s' %s", memName, group.Name, userAddedByName)
			activity := &models.Activity{GroupId: input.GroupId, ActivityDescription: activityDesc}
			if err := tx.Create(activity).Error; err != nil {
				tx.Rollback()
				return err
			}

		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func() {

		var members []string
		for _, member := range input.Members {
			var user models.User
			if err := config.GetDB().Where("id=?", member).First(&user).Error; err != nil {
				fmt.Println("Error fetching user details", err)
				continue
			}

			members = append(members, user.EmailId)

		}
		message := fmt.Sprintf("You were added in '%s' %s", group.Name, userAddedByName)
		Maill(members, "Added in group", message)
	}()

	c.JSON(http.StatusOK, gin.H{"success": "Members addes Successfully"})

}

func GetGroupMembers(c *gin.Context) {
	groupId := c.Query("groupId")

	var groups []models.Group
	if err := config.GetDB().Where("group_id = ?", groupId).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	// fmt.Print(groups)

	var users []models.User
	data := []map[string]interface{}{}

	for i := 0; i < len(groups); i++ {

		var user models.User
		if err := config.GetDB().Where("id=?", groups[i].GroupMembers).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
			return
		}
		users = append(users, user)
	}
	for i := 0; i < len(users); i++ {
		userMap := map[string]interface{}{"name": users[i].Name, "emailId": users[i].EmailId, "memberId": users[i].ID}
		data = append(data, userMap)
	}
	c.JSON(http.StatusOK, data)

}

func DeleteGroupMembers(c *gin.Context) {

	var input MembersInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flag := true
	
	transaction := config.GetDB().Begin()

	for i := 0; i < len(input.Members); i++ {
		if CheckBalance(input.GroupId, input.Members[i]) {

			result := transaction.Where("group_members=?", input.Members[i]).Where("group_id=?", input.GroupId).Where("created_by=?", input.UserId).Delete(&models.Group{})
			if result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}

			if result.RowsAffected == 0 {
				transaction.Rollback()
				c.JSON(http.StatusNotFound, gin.H{"message": "Member not found in the group or user not authorized"})
				return
			}

			var groupName string
			var memberName string
			result = transaction.Table("users").Where("id=?", input.Members[i]).Select("name").Scan(&memberName)
			if result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}

			result = transaction.Table("groups").Where("group_id=?", input.GroupId).Select("name").Scan(&groupName)
			if result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
			activityDesc := fmt.Sprintf("%s was removed from group %s.", memberName,groupName)
			activity := &models.Activity{GroupId: input.GroupId, ActivityDescription: activityDesc}
			if err := transaction.Create(activity).Error; err != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
				return
			}

			
		} else {
			flag = false
		}
	}

	transaction.Commit()

	// go func() {
	// 	var groupName string

	// 	if err := config.GetDB().Model(&models.Group{}).Where("group_id=?", input.GroupId).Select("name").Scan(&groupName).Error; err != nil {
	// 		groupName = ""
	// 	} else {
	// 		groupName = "from" + groupName
	// 	}

	// 	var memIdList []string
	// 	for _, v := range memList {
	// 		var memId string
	// 		if err := config.GetDB().Model(&models.User{}).Where("id=?", v).Select("emai_id").Scan(&memId).Error; err != nil {
	// 			fmt.Println("Error fetching name:", err)
	// 		}
	// 		memIdList = append(memIdList, memId)
	// 	}
	// 	message := fmt.Sprintf("You were removed %s", groupName)

	// 	Maill(memIdList, "Remove from Group", message)
	// }()

	if !flag {
		c.JSON(http.StatusOK, gin.H{"success": "Some Members Cannot be Deleted due to their balance in group."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Members Deleted Successfully"})
}
