package controller

import (
	"bytes"
	"encoding/csv"
	"expensebackend/pkg/config"
	"expensebackend/pkg/models"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func CheckBalanceWithValue(groupId int, memberId int) (float64, error) {

	var lend float64
	var borrow float64
	result1 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owes=? AND group_id=?", memberId, groupId).Scan(&borrow)
	if result1.Error != nil {
		return 0, result1.Error
	}
	result2 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owns=? AND group_id=?", memberId, groupId).Scan(&lend)
	if result2.Error != nil {
		return 0, result2.Error
	}

	return lend - borrow, nil
}

const (
	bucketName = "bucketcsvekansh"
	region     = "ap-south-1" // e.g., "us-west-2"
)

func GenerateCsv(c *gin.Context) {
	// groupId, err := strconv.Atoi(c.Param("groupId"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	// 	return
	// }
	GroupId := c.Query("groupid")
    groupId, err := strconv.Atoi(GroupId)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Invalid group ID: %s", GroupId),
        })
        return
    }
	GroupName := c.Query("groupName")
	// userId, err := strconv.Atoi(c.Query("userId"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	// 	return
	// }

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=report.csv")
	var csvBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuffer)
	// csvWriter := csv.NewWriter(c.Writer)

	var groupMemList []models.Group
	result := config.GetDB().Unscoped().Preload("User").Where("group_id=?", groupId).Find(&groupMemList)
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
		if err := config.GetDB().Table("users").Where("id=?", settlement.UserPaid).Select("name").Scan(&userPaidName).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := config.GetDB().Table("users").Where("id=?", settlement.UserReceived).Select("name").Scan(&userRecivedName).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		row := []string{settlement.CreatedAt.Format("02 Jan 2006"), fmt.Sprintf("%s paid %s", userPaidName, userRecivedName), "Payment", strconv.FormatFloat(settlement.Amount, 'f', 2, 64)}

		for _, memberId := range memberIdList {
			if memberId == settlement.UserPaid {
				row = append(row, strconv.FormatFloat(settlement.Amount, 'f', 2, 64))
			} else if memberId == settlement.UserReceived {
				row = append(row, strconv.FormatFloat(-settlement.Amount, 'f', 2, 64))
			} else {
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
		value, err := CheckBalanceWithValue(groupId, memberId)

		if err != nil {
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

	// Create a session with AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Println("Failed to create AWS session", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create AWS session"})
		return
	}
	s3Client := s3.New(sess)
	// Upload CSV file to S3
	objectKey := fmt.Sprintf("%s.csv", GroupName)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(csvBuffer.Bytes()),
		ContentType: aws.String("text/csv"),
	})
	if err != nil {
		log.Println("Failed to upload CSV to S3", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload CSV to S3"})
		return
	}
	// Generate S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, objectKey)
	// Return the URL
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully uploaded to S3",
		"url":     s3URL,
	})
}
