package router

import "github.com/gin-gonic/gin"

func Initialize() {
	// Initialize router	
	router := gin.Default()

	// Initialize routes 	
	initializeRoutes(router)


	//Running API
	router.Run(":8080")
}