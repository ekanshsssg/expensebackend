package controller

import (
	"expensebackend/pkg/config"
	"expensebackend/pkg/models"
	// "fmt"
	"errors"
	"net/http"
	"os"
	"time"

	// "gorm.io/driver/mysql"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	EmailId  string `json:"emailid" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginInput struct {
	EmailId  string `json:"emailid" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{Name: input.Name, EmailId: input.EmailId}
	if err := user.HashPassword(input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := config.GetDB().Create(&user)
	if result.Error != nil {
		if mySQLError, ok := result.Error.(*mysql.MySQLError); ok && mySQLError.Number == 1062 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email ID already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration success"})

}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.GetDB().Where("email_id = ?", input.EmailId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"emailId": user.EmailId,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := generateToken.SignedString([]byte(os.Getenv("MY_SECRET_TOKEN")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to generate token"})
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"userId":  user.ID,
		"name":    user.Name,
		"message": "login succes",
	})

}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}

// func SuccessConnection(c *gin.Context) {
// 	config.GetDB()

// 	user, exists := c.Get("currentUser")
// 	if !exists {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
// 		return
// 	}

// 	currentUser := user.(models.User)
// 	c.JSON(http.StatusOK, gin.H{"user": currentUser})
// }
