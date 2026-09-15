package ds

type OfficeUser struct {
	ID           uint   `gorm:"primaryKey"`
	UserLogin    string `gorm:"type:varchar(50);unique;not null"`
	PasswordHash string `gorm:"type:varchar(100);not null"`
	IsModerator  bool   `gorm:"type:boolean;default:false"`
}