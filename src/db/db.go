package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(buildDSN()), &gorm.Config{})
	if err != nil {
		panic("Could not connect to database.")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic("Could not configure database connection pool.")
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	if err := DB.AutoMigrate(&Universe{}, &Serie{}, &Book{}); err != nil {
		panic("Could not migrate database schema.")
	}

}

func buildDSN() string {
	host := getEnv("POSTGRES_SERVER", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	name := getEnv("POSTGRES_DB", "lexi_users")
	sslMode := getEnv("POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslMode,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
