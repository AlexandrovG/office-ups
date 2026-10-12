package repository

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"office-uninterruptible_power_supply/internal/app/ds"
)

func (r *Repository) GetPublishedUninterruptiblePowerSupplies(minCap, maxCap float64) ([]ds.UninterruptiblePowerSupply, error) {
	var list []ds.UninterruptiblePowerSupply
	tx := r.db.Preload("Likes").Where("status = ?", "опубликован")
	if minCap > 0 {
		tx = tx.Where("capacity >= ?", minCap)
	}
	if maxCap > 0 {
		tx = tx.Where("capacity <= ?", maxCap)
	}
	err := tx.Find(&list).Error
	return list, err
}

func (r *Repository) GetUninterruptiblePowerSupplyByID(id uint) (*ds.UninterruptiblePowerSupply, error) {
	var ups ds.UninterruptiblePowerSupply
	err := r.db.Preload("Likes").Where("id = ? AND status != ?", id, "удален").First(&ups).Error
	if err != nil {
		return nil, err
	}
	return &ups, nil
}

func (r *Repository) GetNextPublishedUninterruptiblePowerSupply(currentID uint) (*ds.UninterruptiblePowerSupply, error) {
	var list []ds.UninterruptiblePowerSupply
	err := r.db.Preload("Likes").Where("status = ?", "опубликован").Order("id asc").Find(&list).Error
	if err != nil || len(list) == 0 {
		return nil, err
	}
	for i, u := range list {
		if u.ID == currentID && i+1 < len(list) {
			return &list[i+1], nil
		}
	}
	return &list[0], nil
}

func (r *Repository) GetUserDraft(userID uint) (*ds.UninterruptiblePowerSupply, error) {
	var ups ds.UninterruptiblePowerSupply
	err := r.db.Preload("Likes").Where("creator_id = ? AND status = ?", userID, "черновик").First(&ups).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ups, nil
}

func (r *Repository) CreateUninterruptiblePowerSupplyDraft(ups *ds.UninterruptiblePowerSupply) error {
	ups.Status = "черновик"
	ups.CreatedAt = time.Now()
	ups.UpdatedAt = time.Now()
	return r.db.Create(ups).Error
}

func (r *Repository) PublishUninterruptiblePowerSupplyDraft(userID, upsID uint, desc string, capacity, maxCurrent float64) error {
	return r.db.Model(&ds.UninterruptiblePowerSupply{}).
		Where("creator_id = ? AND id = ? AND status = ?", userID, upsID, "черновик").
		Updates(map[string]interface{}{
			"description": desc,
			"capacity":    capacity,
			"max_current": maxCurrent,
			"status":      "опубликован",
			"updated_at":  time.Now(),
		}).Error
}

func (r *Repository) LogicalDeleteUninterruptiblePowerSupplySQL(upsID, userID uint) error {
	query := `UPDATE uninterruptible_power_supplies 
	          SET status = $1, updated_at = $2 
	          WHERE id = $3 AND creator_id = $4 AND status != $1`
	return r.db.Exec(query, "удален", time.Now(), upsID, userID).Error
}

func (r *Repository) LikeUninterruptiblePowerSupply(userID, upsID uint, value int) error {
	if value == 1 {
		like := ds.UninterruptiblePowerSupplyLike{UserID: userID, UpsID: upsID}
		err := r.db.Create(&like).Error
		if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
		return nil
	}
	return r.db.Where("user_id = ? AND ups_id = ?", userID, upsID).
		Delete(&ds.UninterruptiblePowerSupplyLike{}).Error
}

func (r *Repository) UploadUninterruptiblePowerSupplyImage(upsID uint, file *multipart.FileHeader) error {
	filename := fmt.Sprintf("ups_img_%d_%d", upsID, time.Now().UnixNano())
	return r.uploadAndSaveUninterruptiblePowerSupply(upsID, file, filename, "image_url")
}

func (r *Repository) UploadUninterruptiblePowerSupplyVideo(upsID uint, file *multipart.FileHeader) error {
	filename := fmt.Sprintf("ups_vid_%d_%d", upsID, time.Now().UnixNano())
	return r.uploadAndSaveUninterruptiblePowerSupply(upsID, file, filename, "video_url")
}

func (r *Repository) uploadAndSaveUninterruptiblePowerSupply(upsID uint, file *multipart.FileHeader, filename, column string) error {
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = r.minio.PutObject(
		context.Background(),
		r.bucketName,
		filename,
		f,
		file.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://localhost:9000/%s/%s", r.bucketName, filename)
	return r.db.Model(&ds.UninterruptiblePowerSupply{}).
		Where("id = ?", upsID).
		Update(column, url).Error
}

func (r *Repository) CreateUser(login, passwordHash string) (*ds.OfficeUser, error) {
	user := ds.OfficeUser{UserLogin: login, PasswordHash: passwordHash}
	err := r.db.Create(&user).Error
	return &user, err
}