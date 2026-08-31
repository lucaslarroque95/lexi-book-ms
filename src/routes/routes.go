package routes

import (
	"lexi/books/middlewares"
	"lexi/books/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, bookHandler *BookHandler, universeHandler *UniverseHandler, serieHandler *SerieHandler, keys utils.Keys) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	user := router.Group("/")
	user.Use(middlewares.Authenticate(keys))

	user.GET("/books", bookHandler.GetBooks)
	user.GET("/books/name/:name", bookHandler.GetBookByName)
	user.GET("/books/:id", bookHandler.GetBookByID)
	user.GET("/books/:id/download-url", bookHandler.GetBookDownloadURL)
	user.POST("/books", bookHandler.CreateBook)
	user.PUT("/books/:id", bookHandler.UpdateBook)
	user.DELETE("/books/:id", bookHandler.DeleteBook)

	user.GET("/universes", universeHandler.GetUniverses)
	user.GET("/universes/:id", universeHandler.GetUniverseByID)
	user.POST("/universes", universeHandler.CreateUniverse)
	user.PUT("/universes/:id", universeHandler.UpdateUniverse)
	user.DELETE("/universes/:id", universeHandler.DeleteUniverse)

	user.GET("/series", serieHandler.GetSeries)
	user.GET("/series/:id", serieHandler.GetSerieByID)
	user.POST("/series", serieHandler.CreateSerie)
	user.PUT("/series/:id", serieHandler.UpdateSerie)
	user.DELETE("/series/:id", serieHandler.DeleteSerie)
}
