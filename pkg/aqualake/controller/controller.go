package controller

import "github.com/gin-gonic/gin"

func SetUpRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/function/:name", func(ctx *gin.Context) {

	})

	router.PUT("/function", func(ctx *gin.Context) {

	})

	router.DELETE("/function/:name", func(ctx *gin.Context) {

	})

	return router
}