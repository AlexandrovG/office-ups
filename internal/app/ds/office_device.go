package ds

import (
	"database/sql"
	"time"
)

type OfficeDevice struct {
	ID                uint           `gorm:"primaryKey"`
	DeviceName        string         `gorm:"type:varchar(100);not null"`
	DeviceDescription string         `gorm:"type:text"`
	DeviceStatus      string         `gorm:"type:varchar(20);not null;default:'черновик'"`
	ImageURL          sql.NullString `gorm:"type:varchar(255)"`
	VideoURL          sql.NullString `gorm:"type:varchar(255)"`
	ActivePowerWatts  int            `gorm:"not null;default:0"`
	PowerFactor       float64        `gorm:"type:numeric(3,2);not null;default:0.70"`
	CreatedAt         time.Time      `gorm:"not null"`
	UpdatedAt         time.Time
	CreatorID         uint           `gorm:"not null"`

	Creator OfficeUser `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT"`
}