package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "office-uninterruptible_power_supply/docs"

	"office-uninterruptible_power_supply/internal/app/dsn"
	"office-uninterruptible_power_supply/internal/app/handler"
	"office-uninterruptible_power_supply/internal/app/redis"
	"office-uninterruptible_power_supply/internal/app/repository"
)

// @title Office UPS API
// @version 1.0
// @description API для управления ИБП в офисе
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	redisClient, err := redis.New(context.Background(), os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PASSWORD"), redisPort)
	if err != nil {
		log.Fatalf("Ошибка инициализации Redis: %v", err)
	}
	defer redisClient.Close()

	h := handler.NewHandler(repo, redisClient)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.POST("/auth/register", h.RegisterAPI)
		api.POST("/auth/login", h.LoginAPI)

		ups := api.Group("/uninterruptible-power-supplies")
		{
			ups.GET("", h.GetUninterruptiblePowerSuppliesListAPI)
			ups.GET("/reels", h.GetUninterruptiblePowerSuppliesReelsAPI)

			auth := ups.Group("")
			auth.Use(h.WithAuthCheck())
			{
				auth.GET("/draft", h.GetUninterruptiblePowerSupplyDraftAPI)
				auth.POST("", h.CreateUninterruptiblePowerSupplyAPI)
				auth.PUT("/publish", h.PublishUninterruptiblePowerSupplyAPI)
				auth.DELETE("/:id", h.DeleteUninterruptiblePowerSupplyAPI)
				auth.POST("/:id/like", h.LikeUninterruptiblePowerSupplyAPI)
			}
		}

		authGroup := api.Group("/auth")
		authGroup.Use(h.WithAuthCheck())
		{
			authGroup.POST("/logout", h.LogoutAPI)
		}
	}

	r.Run(":8080")
}