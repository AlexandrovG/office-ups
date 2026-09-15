package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"office-ups/internal/app/dsn"
	"office-ups/internal/app/handler"
	"office-ups/internal/app/repository"
)

func main() {
	_ = godotenv.Load()

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		log.Fatalf("Ошибка соединения с БД: %v", err)
	}

	minioURL := os.Getenv("MINIO_URL")
	if minioURL == "" {
		minioURL = "http://localhost:9000/ups-media/"
	}

	h := handler.NewHandler(repo, minioURL)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/office-devices/feed", h.GetFeed)
	r.GET("/office-devices/draft", h.GetDraft)
	r.GET("/office-devices/grid", h.GetGrid)

	r.POST("/office-devices/draft/create", h.PostCreateDraft)
	r.POST("/office-devices/draft/publish", h.PostPublishDraft)
	r.POST("/office-devices/delete", h.PostDeleteDevice)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/office-devices/grid")
	})

	r.Run(":8080")
}