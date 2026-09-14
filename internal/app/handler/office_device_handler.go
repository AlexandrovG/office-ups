package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"office-ups/internal/app/repository"
)

type OfficeDeviceViewDTO struct {
	ID                int
	DeviceName        string
	DeviceDescription string
	ActivePowerWatts  int
	PowerFactor       float64
	LikesCount        int
	ImageURL          string
	VideoURL          string
}

type OfficeDeviceHandler struct {
	repo     *repository.OfficeDeviceRepository
	minioURL string
}

func NewOfficeDeviceHandler(repo *repository.OfficeDeviceRepository, minioURL string) *OfficeDeviceHandler {
	return &OfficeDeviceHandler{
		repo:     repo,
		minioURL: minioURL,
	}
}

func (h *OfficeDeviceHandler) toDTO(d repository.OfficeDevice) OfficeDeviceViewDTO {
	return OfficeDeviceViewDTO{
		ID:                d.ID,
		DeviceName:        d.DeviceName,
		DeviceDescription: d.DeviceDescription,
		ActivePowerWatts:  d.ActivePowerWatts,
		PowerFactor:       d.PowerFactor,
		LikesCount:        len(d.LikedUserIDs),
		ImageURL:          fmt.Sprintf("%s%s", h.minioURL, d.ImageKey),
		VideoURL:          fmt.Sprintf("%s%s", h.minioURL, d.VideoKey),
	}
}

func (h *OfficeDeviceHandler) RenderFeed(c *gin.Context) {
	idStr := c.Query("device_id")
	nextStr := c.Query("next")

	var device repository.OfficeDevice
	var err error

	if idStr == "" {
		published := h.repo.GetPublishedDevices(0)
		if len(published) > 0 {
			device = published[0]
		}
	} else {
		id, _ := strconv.Atoi(idStr)
		if nextStr == "true" {
			device = h.repo.GetNextPublishedDevice(id)
		} else {
			device, err = h.repo.GetDeviceByID(id)
			if err != nil {
				c.String(http.StatusNotFound, "Оборудование не найдено")
				return
			}
		}
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"Device": h.toDTO(device),
	})
}

func (h *OfficeDeviceHandler) RenderDraft(c *gin.Context) {
	draft, err := h.repo.GetDraftDevice()
	if err != nil {
		c.String(http.StatusNotFound, "Черновик не найден")
		return
	}

	c.HTML(http.StatusOK, "draft.html", gin.H{
		"Device": h.toDTO(draft),
	})
}

func (h *OfficeDeviceHandler) RenderGrid(c *gin.Context) {
	powerQuery := c.Query("min_power_watts")
	minWatts := 0
	if powerQuery != "" {
		minWatts, _ = strconv.Atoi(powerQuery)
	}

	rawDevices := h.repo.GetPublishedDevices(minWatts)
	var dtos []OfficeDeviceViewDTO
	for _, d := range rawDevices {
		dtos = append(dtos, h.toDTO(d))
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"Devices":        dtos,
		"MinPowerFilter": powerQuery,
	})
}