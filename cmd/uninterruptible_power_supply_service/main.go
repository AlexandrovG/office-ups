package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"office-uninterruptible_power_supply/internal/app/dsn"
	"office-uninterruptible_power_supply/internal/app/handler"
	"office-uninterruptible_power_supply/internal/app/repository"
)

func main() {
	_ = godotenv.Load()

	repo, err := repository.New(&repository.RepositorySettings{
		PostgresDSN:     dsn.FromEnv(),
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
		UseSSL:          false,
	})
	if err != nil {
		log.Fatalf("Ошибка инициализации репозитория: %v", err)
	}

	h := handler.NewHandler(repo)

	r := gin.Default()

	api := r.Group("/api")
	{
		ups := api.Group("/uninterruptible-power-supplies")
		{
			ups.GET("", h.GetUninterruptiblePowerSuppliesListAPI)
			ups.GET("/reels", h.GetUninterruptiblePowerSuppliesReelsAPI)
			ups.GET("/draft", h.GetUninterruptiblePowerSupplyDraftAPI)
			ups.POST("", h.CreateUninterruptiblePowerSupplyAPI)
			ups.PUT("/publish", h.PublishUninterruptiblePowerSupplyAPI)
			ups.DELETE("/:id", h.DeleteUninterruptiblePowerSupplyAPI)
			ups.POST("/:id/like", h.LikeUninterruptiblePowerSupplyAPI)
		}

		api.POST("/auth/register", h.RegisterAPI)
		api.POST("/auth/login", h.LoginAPI)
		api.POST("/auth/logout", h.LogoutAPI)
	}

	r.Run(":8080")
}