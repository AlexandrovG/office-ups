package ds

import "time"

type OfficeDeviceLike struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_device"`
	DeviceID  uint      `gorm:"not null;uniqueIndex:idx_user_device"`
	CreatedAt time.Time `gorm:"not null"`

	User   OfficeUser   `gorm:"foreignKey:UserID;constraint:OnDelete:RESTRICT"`
	Device OfficeDevice `gorm:"foreignKey:DeviceID;constraint:OnDelete:RESTRICT"`
}