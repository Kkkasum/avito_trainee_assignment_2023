package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "avito_2023/docs"
	sh "avito_2023/internal/segment/handler"
	sr "avito_2023/internal/segment/repo"
	uh "avito_2023/internal/user/handler"
	ur "avito_2023/internal/user/repo"
)

const (
	envPath = ".env"

	addr = ":8080"
)

// @title Avito Trainee Assignment 2023
// @version 1.0
// @description User Segmentation Service

// @host localhost:8080
// @BasePath /

func main() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	segmentRepo := sr.NewRepo(db)
	segmentHandler := sh.NewHandler(segmentRepo)
	sh.Route(r, segmentHandler)

	userRepo := ur.NewRepo(db)
	userHandler := uh.NewHandler(userRepo)
	uh.Route(r, userHandler)

	log.Println("Starting app...")
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %s", err)
	}
}
