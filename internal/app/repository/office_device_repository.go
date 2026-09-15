package repository

import (
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"office-ups/internal/app/ds"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) GetPublishedDevices(minWatts int) ([]ds.OfficeDevice, error) {
	var devices []ds.OfficeDevice
	tx := r.db.Where("device_status = ?", "опубликован")
	if minWatts > 0 {
		tx = tx.Where("active_power_watts >= ?", minWatts)
	}
	err := tx.Find(&devices).Error
	return devices, err
}

func (r *Repository) GetDeviceByID(id uint) (*ds.OfficeDevice, error) {
	var dev ds.OfficeDevice
	err := r.db.Where("id = ? AND device_status != ?", id, "удален").First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

func (r *Repository) GetNextPublishedDevice(currentID uint) (*ds.OfficeDevice, error) {
	var devices []ds.OfficeDevice
	err := r.db.Where("device_status = ?", "опубликован").Order("id asc").Find(&devices).Error
	if err != nil || len(devices) == 0 {
		return nil, err
	}
	for i, d := range devices {
		if d.ID == currentID && i+1 < len(devices) {
			return &devices[i+1], nil
		}
	}
	return &devices[0], nil
}

func (r *Repository) GetUserDraft(userID uint) (*ds.OfficeDevice, error) {
	var dev ds.OfficeDevice
	err := r.db.Where("creator_id = ? AND device_status = ?", userID, "черновик").First(&dev).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dev, nil
}

func (r *Repository) CreateDraft(dev *ds.OfficeDevice) error {
	dev.DeviceStatus = "черновик"
	dev.CreatedAt = time.Now()
	dev.UpdatedAt = time.Now()
	return r.db.Create(dev).Error
}

func (r *Repository) PublishDraft(userID uint, desc string, watts int, pf float64) error {
	return r.db.Model(&ds.OfficeDevice{}).
		Where("creator_id = ? AND device_status = ?", userID, "черновик").
		Updates(map[string]interface{}{
			"device_description": desc,
			"active_power_watts": watts,
			"power_factor":       pf,
			"device_status":      "опубликован",
			"updated_at":         time.Now(),
		}).Error
}

func (r *Repository) LogicalDeleteSQL(deviceID uint) error {
	query := "UPDATE office_devices SET device_status = $1, updated_at = $2 WHERE id = $3"
	return r.db.Exec(query, "удален", time.Now(), deviceID).Error
}

func (r *Repository) GetDeviceLikesCount(deviceID uint) int64 {
	var count int64
	r.db.Model(&ds.OfficeDeviceLike{}).Where("device_id = ?", deviceID).Count(&count)
	return count
}