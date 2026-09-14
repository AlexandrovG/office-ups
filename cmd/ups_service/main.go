package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"office-ups/internal/app/handler"
	"office-ups/internal/app/repository"
)

func main() {
	repo := repository.NewOfficeDeviceRepository()
	minioBaseURL := "http://localhost:9000/ups-media/"
	h := handler.NewOfficeDeviceHandler(repo, minioBaseURL)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/office-devices/feed", h.RenderFeed)
	r.GET("/office-devices/draft", h.RenderDraft)
	r.GET("/office-devices/grid", h.RenderGrid)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/office-devices/feed")
	})

	log.Println("Сервер доступен по адресу: http://localhost:8080")
	r.Run(":8080")
}