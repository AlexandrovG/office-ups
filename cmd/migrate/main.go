package main

import (
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"office-uninterruptible_power_supply/internal/app/ds"
	"office-uninterruptible_power_supply/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}

	err = db.AutoMigrate(
		&ds.OfficeUser{},
		&ds.UninterruptiblePowerSupply{},
		&ds.UninterruptiblePowerSupplyLike{},
	)
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	var count int64
	db.Model(&ds.OfficeUser{}).Where("id = ?", 1).Count(&count)
	if count == 0 {
		db.Create(&ds.OfficeUser{ID: 1, UserLogin: "engineer_ivan", PasswordHash: "hash1", IsModerator: false})
	}

	log.Println("Таблицы успешно созданы")
}