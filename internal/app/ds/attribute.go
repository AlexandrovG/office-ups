package ds

type Attribute struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100);not null" json:"name"`
}

type UpsAttribute struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UpsID       uint   `gorm:"not null" json:"ups_id"`
	AttributeID uint   `gorm:"not null" json:"attribute_id"`
	Value       string `gorm:"type:varchar(255)" json:"value"`

	Attribute Attribute `gorm:"foreignKey:AttributeID" json:"attribute"`
}