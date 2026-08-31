package main

import (
	"log"

	"lexi/books/db"
	_ "lexi/books/docs"
	"lexi/books/queue"
	"lexi/books/repositories"
	"lexi/books/routes"
	"lexi/books/service"
	"lexi/books/storage"
	"lexi/books/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title        Lexi Book Service API
// @version      1.0
// @description  Manages books, series, and universes for the Lexi platform.
// @BasePath     /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Raw JWT issued by lexi-users-ms's /login (no "Bearer " prefix).
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system environment variables")
	}

	db.InitDB()

	keys := utils.Keys{}
	if err := keys.LoadPublicKey("keys/public.pem"); err != nil {
		log.Fatalf("failed to load public key: %v", err)
	}

	objectStorage := storage.NewObjectStorageFromEnv()

	publisher, err := queue.NewPublisherFromEnv()
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	bookRepository := repositories.NewPostgresBookRepository(db.DB)
	serieRepository := repositories.NewPostgresSerieRepository(db.DB)
	universeRepository := repositories.NewPostgresUniverseRepository(db.DB)

	bookService := service.NewBookService(bookRepository, objectStorage, publisher)
	bookHandler := routes.NewBookHandler(bookService)

	serieService := service.NewSerieService(serieRepository, bookRepository)
	serieHandler := routes.NewSerieHandler(serieService)

	universeService := service.NewUniverseService(universeRepository, serieRepository)
	universeHandler := routes.NewUniverseHandler(universeService)

	server := gin.Default()
	routes.RegisterRoutes(server, bookHandler, universeHandler, serieHandler, keys)

	server.Run(":8080")
}
