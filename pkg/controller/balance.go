package controller

import (
	"expensebackend/pkg/config"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

func CheckBalance(groupId int, memberId int) bool {

	var lend float64
	var borrow float64
	result1 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owes=? AND group_id=?", memberId, groupId).Scan(&borrow)
	if result1.Error != nil {
		return false
	}
	result2 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owns=? AND group_id=?", memberId, groupId).Scan(&lend)
	if result2.Error != nil {
		return false
	}

	if lend-borrow == 0 {
		return true
	}
	return false
}

type Balance struct {
	UserId  int     `json:"userId"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

func GetBalances(c *gin.Context) {
	groupId, err := strconv.Atoi(c.Query("groupId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_userId, err := strconv.Atoi(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var balances []Balance
	rows, err := config.GetDB().Raw(`select user_who_owes as user_id, u.name,SUM(amount) as balance from debts d JOIN users u on u.id=d.user_who_owes WHERE group_id=? AND user_who_owns=? GROUP BY user_who_owes`, groupId, _userId).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer rows.Close()

	for rows.Next() {
		var balance Balance
		if err := rows.Scan(&balance.UserId, &balance.Name, &balance.Balance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if balance.Balance != 0 {
			balances = append(balances, balance)
		}
	}

	rows, err = config.GetDB().Raw(`select user_who_owns as user_id, u.name,SUM(amount) as balance from debts d JOIN users u on u.id=d.user_who_owns WHERE group_id=? AND user_who_owes=? GROUP BY user_who_owns`, groupId, _userId).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userId int
		var name string
		var balance float64
		if err := rows.Scan(&userId, &name, &balance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if balance != 0 {

			found := false

			for i, v := range balances {
				if v.UserId == userId {
					balances[i].Balance -= balance
					found = true
					break
				}
			}

			if !found {
				balances = append(balances, Balance{
					UserId:  userId,
					Name:    name,
					Balance: -balance,
				})
			}
		}
	}

	if len(balances) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"balances": balances})
}

func GetOverallBalance(c *gin.Context) {
	_userId, err := strconv.Atoi(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var lend float64
	var borrow float64
	result1 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owes=?", _userId).Scan(&borrow)
	if result1.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result1.Error.Error()})
		return
	}
	result2 := config.GetDB().Table("debts").Select("coalesce(sum(amount),0.0)").Where("user_who_owns=?", _userId).Scan(&lend)
	if result2.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result2.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": float64(lend - borrow)})

}