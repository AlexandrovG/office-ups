package ds

type UninterruptiblePowerSupplyLike struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;uniqueIndex:idx_user_ups"`
	UpsID  uint `gorm:"not null;uniqueIndex:idx_user_ups"`

	User OfficeUser                `gorm:"foreignKey:UserID"`
	Ups  UninterruptiblePowerSupply `gorm:"foreignKey:UpsID"`
}