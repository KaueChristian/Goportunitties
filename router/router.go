package router

import "github.com/gin-gonic/gin"

func Initialize() {
	router := gin.Default()
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})
	//Running API
	router.Run(":8080")
}