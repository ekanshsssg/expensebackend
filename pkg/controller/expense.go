package controller

import (
	"expensebackend/pkg/config"
	"expensebackend/pkg/models"
	"fmt"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"

	_ "github.com/jinzhu/gorm"
	"gorm.io/gorm"
)

type AddExpenseInput struct {
	GroupId        int     `json:"group_id" binding:"required"`
	Category       string  `json:"category" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	Description    string  `json:"description" binding:"required"`
	PaidBy         int     `json:"paid_by" binding:"required"`
	ExpenseMembers []int   `json:"expense_members" binding:"required"`
}

func AddExpense(c *gin.Context) {

	var input AddExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Print(input)

	transaction := config.GetDB().Begin()

	enterExpense := &models.Expense{GroupId: input.GroupId, Category: input.Category, Amount: input.Amount, ExpenseDescription: input.Description, PaidBy: input.PaidBy}

	result := transaction.Create(enterExpense)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	expenseSplitAmount := input.Amount / float64(len(input.ExpenseMembers))

	for _, mem := range input.ExpenseMembers {
		enterExpenseMember := &models.Expensemember{ExpenseId: enterExpense.ExpenseId, GroupId: input.GroupId, ExpenseMember: mem, Amount: expenseSplitAmount}
		result := transaction.Create(enterExpenseMember)
		if result.Error != nil {
			transaction.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
	}

	for _, member := range input.ExpenseMembers {

		if input.PaidBy != member {

			result1 := transaction.Model(&models.Debt{}).Where("group_id=? AND user_who_owns=? AND user_who_owes=?", input.GroupId, input.PaidBy, member).Update("amount", gorm.Expr("amount + ?", expenseSplitAmount))

			if result1.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
			if result1.RowsAffected == 0 {
				insertIntoDebt := &models.Debt{GroupId: input.GroupId, UserWhoOwns: input.PaidBy, UserWhoOwes: member, Amount: expenseSplitAmount}
				result2 := transaction.Create(insertIntoDebt)
				if result2.Error != nil {
					transaction.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
					return
				}
			}
		}
	}

	var userPaidName string
	var groupName string
	result = transaction.Model(&models.User{}).Where("id=?", input.PaidBy).Select("name").Scan(&userPaidName)
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

	activityDesc := fmt.Sprintf("%s added an expense '%s' of amount %s in '%s' ", userPaidName, input.Description, strconv.FormatFloat(input.Amount, 'f', 2, 64), groupName)
	activity := &models.Activity{GroupId: input.GroupId, ActivityDescription: activityDesc}
	if err := transaction.Create(activity).Error; err != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	transaction.Commit()

	var group models.Group
	var paidBy models.User
	// var groupName string

	if err := config.GetDB().Where("group_id=?", input.GroupId).First(&group).Error; err != nil {
		groupName = ""
	} else {
		groupName = group.Name
	}
	if err := config.GetDB().Where("id=?", input.PaidBy).First(&paidBy).Error; err == nil {

		go func() {

			var members []string
			for _, member := range input.ExpenseMembers {
				var user models.User
				if err := config.GetDB().Where("id=?", member).First(&user).Error; err != nil {
					fmt.Println("Error fetching user details", err)
					continue
				}

				if user.ID != paidBy.ID {
					members = append(members, user.EmailId)
				}

			}
			message := fmt.Sprintf("%s added '%s' in '%s' . You owe ₹ %f", paidBy.Name, enterExpense.ExpenseDescription, groupName, expenseSplitAmount)
			Maill(members, "Expense added", message)
		}()
	} else {
		fmt.Println("Error fetching user")
	}

	c.JSON(http.StatusOK, gin.H{"expenseId": enterExpense.ExpenseId})

}

func GetExpenses(c *gin.Context) {
	_userId, err := strconv.Atoi(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	groupId, err := strconv.Atoi(c.Query("groupId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var allExpenses []models.Expense
	result := config.GetDB().Preload("User").Where("group_id=?", groupId).Order("created_at desc").Find(&allExpenses)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	var response []map[string]interface{}
	for _, expense := range allExpenses {
		var amount float64

		if expense.PaidBy == _userId {
			result := config.GetDB().Table("expenseMembers").Select("Coalesce(SUM(amount),0)").Where("expense_id=? AND expense_member != ?", expense.ExpenseId, _userId).Scan(&amount)
			if result.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
		} else {
			result := config.GetDB().Table("expenseMembers").Select("Coalesce(SUM(amount),0)").Where("expense_id=? AND expense_member = ?", expense.ExpenseId, _userId).Scan(&amount)
			amount = -amount
			if result.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
		}
		// print(expense.CreatedAt.Format("02 Jul"));
		response = append(response, map[string]interface{}{
			"expense_id":   expense.ExpenseId,
			"category":     expense.Category,
			"amount":       expense.Amount,
			"description":  expense.ExpenseDescription,
			"paid_by":      expense.PaidBy,
			"paid_by_name": expense.User.Name,
			"userAmount":   amount,
			"created_at":   expense.CreatedAt.Format("02 Jan"),
		})
	}

	print(response)
	if len(response) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expenses": response,
	})

}