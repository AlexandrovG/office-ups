package ds

type OfficeUser struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	UserLogin    string `gorm:"type:varchar(50);unique;not null" json:"login"`
	PasswordHash string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator  bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}