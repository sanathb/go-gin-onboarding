package main

import (
	"myConstants"
	"myDatabase"
	"myMiddleware"
	"onboarding"

	"github.com/gin-gonic/gin"
)

func main() {

	var db = myDatabase.DBConnect()

	router := gin.Default()

	router.Use(myMiddleware.RestrictInputContent)
	router.Use(myMiddleware.CheckToken(db))
	router.Use(myMiddleware.RequestLoggerMiddleware())
	//router.Use(myMiddleware.JsonSchemaValidator())

	router.POST("/signup", onboarding.Signup(db))
	router.POST("/login", onboarding.Login(db))
	router.POST("/logout", onboarding.Logout(db))

	defer db.Close()

	//Listen and serve
	router.Run(myConstants.ServerIP + ":" + myConstants.ServerPort)

}
