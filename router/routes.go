package router

import (
	"github.com/KaueChristian/Goportunitties/handler"
	"github.com/gin-gonic/gin"
)

func initializeRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/opening", handler.CreatOpeningHandler)
		v1.DELETE("/opening", handler.DeleteOpeningHandler)
		v1.POST("/opening", handler.ShowOpeningHandler)
		v1.PUT("/opening", handler.UpdateOpeningHandler)
	}
}