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
	log.Println("Таблицы успешно созданы")
}