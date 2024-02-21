package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func sendError(ctx *gin.Context, code int, msg string) {
	ctx.Header("content-type", "aplicattion/json")
	ctx.JSON(code, gin.H{
		"MSG":       msg,
		"ERRORCODE": code,
	})
}

func sendSucess(ctx *gin.Context, op string, data interface{}) {
	ctx.Header("content-type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"msg" : fmt.Sprintf("operation from handler: %s successfull", op),
		"data" : data,
	})
}
