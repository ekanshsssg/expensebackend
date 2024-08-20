package main

import (
    "os"
    "expensebackend/pkg/config"
    "expensebackend/pkg/routes"
)

func init(){
    config.LoadEnvVariables()
    config.ConnectDatabase()
}

func main() {
	// r := gin.Default()
	
    // config.GetDB().AutoMigrate(&models.User{})

    r := routes.SetupRouter()
    r.Run(":"+os.Getenv("PORT"))
}
