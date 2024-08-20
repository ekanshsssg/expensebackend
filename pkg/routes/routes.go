package routes

import (
	"expensebackend/pkg/controller"
	"expensebackend/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.POST("/auth/register", controller.Register)
	router.POST("/auth/login", controller.Login)
	router.POST("/auth/logout", controller.Logout)

	// router.GET("/", middlewares.CheckAuth, controller.SuccessConnection)
	router.POST("/group/create-group",controller.CreateGroup);
	router.GET("/group/:userId",middlewares.CheckAuth,controller.GetGroups);
	router.GET("/group/search-members",middlewares.CheckAuth,controller.GetMember);
	router.POST("/group/add-members",controller.AddMembers);
	router.GET("/group/get-members",middlewares.CheckAuth,controller.GetGroupMembers);
	router.POST("/group/delete-members",middlewares.CheckAuth,controller.DeleteGroupMembers);

	router.POST("/add-expense",middlewares.CheckAuth,controller.AddExpense);
	router.GET("/get-expense",controller.GetExpenses);
	router.GET("/get-balance",controller.GetBalances);
	router.GET("/get-overall-balance",controller.GetOverallBalance);
	router.POST("/add-settlement",controller.AddSettlement);

	router.GET("/activity",controller.GetActivity)
	router.GET("/get-csv",controller.GenerateCsv)

	return router
}
