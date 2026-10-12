package ds

import (
	"database/sql"
	"time"
)

type UninterruptiblePowerSupply struct {
	ID          uint           `gorm:"primaryKey"`
	UpsName     string         `gorm:"type:varchar(100);not null"`
	Description string         `gorm:"type:text"`
	Status      string         `gorm:"type:varchar(20);not null;default:'черновик'"`
	ImageURL    sql.NullString `gorm:"type:varchar(255)"`
	VideoURL    sql.NullString `gorm:"type:varchar(255)"`
	Capacity    float64        `gorm:"type:numeric(10,2)"`
	MaxCurrent  float64        `gorm:"type:numeric(10,2)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatorID   uint           `gorm:"not null"`

	Creator OfficeUser                       `gorm:"foreignKey:CreatorID"`
	Likes   []UninterruptiblePowerSupplyLike `gorm:"foreignKey:UpsID"`
}