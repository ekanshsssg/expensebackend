package controller

import (
	"expensebackend/pkg/config"
	"fmt"

	"encoding/csv"
	"expensebackend/pkg/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CheckBalanceWithValue(groupId int, memberId int) (float64,error) {

	var lend float64
	var borrow float64
	result1 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owes=? AND group_id=?", memberId, groupId).Scan(&borrow)
	if result1.Error != nil {
		return 0,result1.Error
	}
	result2 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owns=? AND group_id=?", memberId, groupId).Scan(&lend)
	if result2.Error != nil {
		return 0,result2.Error
	}

	
	return lend-borrow,nil
}

func GenerateCsv(c *gin.Context) {
	groupId, err := strconv.Atoi(c.Query("groupId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	// userId, err := strconv.Atoi(c.Query("userId"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	// 	return
	// }

	c.Header("Content-Type", "text/csv")
	c.Header("Context-Disposition", "attachment;filename=report.csv")
	csvWriter := csv.NewWriter(c.Writer)

	var groupMemList []models.Group
	result := config.GetDB().Preload("User").Where("group_id=?", groupId).Find(&groupMemList)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// print(result)

	var allExpenses []models.Expense
	result = config.GetDB().Preload("User").Where("group_id=?", groupId).Find(&allExpenses)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	headers := []string{"Date", "Description", "Category", "Amount"}
	var memberIdList []int
	for _, v := range groupMemList {
		memberIdList = append(memberIdList, v.GroupMembers)
		headers = append(headers, v.User.Name)
	}

	if err := csvWriter.Write(headers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// print("--------------")
	// print(len(allExpenses))
	// c.JSON(http.StatusOK, gin.H{"expense": allExpenses, "groups": groupMemList})

	// var rows [][]string

	// var response []map[string]interface{}
	for _, expense := range allExpenses {

		row := []string{expense.CreatedAt.Format("02 Jan 2006"), expense.ExpenseDescription, "Expense", strconv.FormatFloat(expense.Amount, 'f', 2, 64)}

		var expenseMemList []models.Expensemember

		result := config.GetDB().Where("expense_id=?", expense.ExpenseId).Find(&expenseMemList)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		expenseMemMap := make(map[int]float64)
		for _, expenseMember := range expenseMemList {
			if expenseMember.ExpenseMember == expense.PaidBy {
				expenseMemMap[expenseMember.ExpenseMember] = expenseMember.Amount
			} else {
				expenseMemMap[expenseMember.ExpenseMember] = -expenseMember.Amount
			}
		}

		for _, memberId := range memberIdList {
			if amount, exists := expenseMemMap[memberId]; exists {
				row = append(row, strconv.FormatFloat(amount, 'f', 2, 64))
			} else {
				row = append(row, strconv.FormatFloat(0, 'f', 2, 64))
			}
		}

		if err := csvWriter.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		// rows = append(rows, row)
	}
	
	if err := csvWriter.Write([]string{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	var settlements []models.Settlement
	if err := config.GetDB().Where("group_id=?", groupId).Find(&settlements).Order("created_at desc").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, settlement := range settlements {

		var userPaidName string
		var userRecivedName string
		if err := config.GetDB().Table("users").Where("id=?", settlement.UserPaid).Select("name").Scan(&userPaidName).Error;err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		if err := config.GetDB().Table("users").Where("id=?", settlement.UserReceived).Select("name").Scan(&userRecivedName).Error;err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		row := []string{settlement.CreatedAt.Format("02 Jan 2006"), fmt.Sprintf("%s paid %s",userPaidName,userRecivedName), "Payment", strconv.FormatFloat(settlement.Amount, 'f', 2, 64)}

		for _, memberId := range memberIdList {
			if  memberId == settlement.UserPaid{
				row = append(row, strconv.FormatFloat(settlement.Amount, 'f', 2, 64))
			} else if  memberId == settlement.UserReceived{
				row = append(row, strconv.FormatFloat(-settlement.Amount, 'f', 2, 64))
			}else{
				row = append(row, strconv.FormatFloat(0, 'f', 2, 64))
			}
		}

		if err := csvWriter.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
	}

	if err := csvWriter.Write([]string{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}


	row := []string{"", "Total balance", "", ""}
	for _, memberId := range memberIdList {
		value,err := CheckBalanceWithValue(groupId,memberId)

		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
		}

		row = append(row, strconv.FormatFloat(value, 'f', 2, 64))
	}

	if err := csvWriter.Write(row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error flushing csv data"})
		return
	}

	// c.JSON(http.StatusOK,gin.H{"rows":rows})
}
