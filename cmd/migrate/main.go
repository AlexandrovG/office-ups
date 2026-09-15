package main

import (
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"office-ups/internal/app/ds"
	"office-ups/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	err = db.AutoMigrate(
		&ds.OfficeUser{},
		&ds.OfficeDevice{},
		&ds.OfficeDeviceLike{},
	)
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}
	log.Println("Таблицы успешно созданы")
}