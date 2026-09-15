package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"office-ups/internal/app/ds"
	"office-ups/internal/app/repository"
)

type DeviceViewDTO struct {
	ID                uint
	DeviceName        string
	DeviceDescription string
	ActivePowerWatts  int
	PowerFactor       float64
	LikesCount        int64
	ImageURL          string
	VideoURL          string
}

type Handler struct {
	repo     *repository.Repository
	minioURL string
}

func NewHandler(r *repository.Repository, minioURL string) *Handler {
	return &Handler{repo: r, minioURL: minioURL}
}

func (h *Handler) toDTO(d *ds.OfficeDevice) DeviceViewDTO {
	img := "/static/img/default_device.png"
	if d.ImageURL.Valid && d.ImageURL.String != "" {
		img = fmt.Sprintf("%s%s", h.minioURL, d.ImageURL.String)
	}

	video := "/static/video/default_preview.mp4"
	if d.VideoURL.Valid && d.VideoURL.String != "" {
		video = fmt.Sprintf("%s%s", h.minioURL, d.VideoURL.String)
	}

	return DeviceViewDTO{
		ID:                d.ID,
		DeviceName:        d.DeviceName,
		DeviceDescription: d.DeviceDescription,
		ActivePowerWatts:  d.ActivePowerWatts,
		PowerFactor:       d.PowerFactor,
		LikesCount:        h.repo.GetDeviceLikesCount(d.ID),
		ImageURL:          img,
		VideoURL:          video,
	}
}

func (h *Handler) GetFeed(c *gin.Context) {
	idStr := c.Query("device_id")
	next := c.Query("next")

	var dev *ds.OfficeDevice
	var err error

	if idStr == "" {
		list, _ := h.repo.GetPublishedDevices(0)
		if len(list) > 0 {
			dev = &list[0]
		}
	} else {
		id, _ := strconv.Atoi(idStr)
		if next == "true" {
			dev, _ = h.repo.GetNextPublishedDevice(uint(id))
		} else {
			dev, err = h.repo.GetDeviceByID(uint(id))
			if err != nil {
				c.String(http.StatusNotFound, "Оборудование удалено или отсутствует")
				return
			}
		}
	}

	if dev == nil {
		c.String(http.StatusOK, "Опубликованное оборудование отсутствует")
		return
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{"Device": h.toDTO(dev)})
}

func (h *Handler) GetDraft(c *gin.Context) {
	userID := uint(1)
	draft, err := h.repo.GetUserDraft(userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	hasDraft := draft != nil
	var dto DeviceViewDTO
	if hasDraft {
		dto = h.toDTO(draft)
	}

	c.HTML(http.StatusOK, "draft.html", gin.H{
		"HasDraft": hasDraft,
		"Device":   dto,
	})
}

func (h *Handler) GetGrid(c *gin.Context) {
	powerStr := c.Query("min_power_watts")
	minWatts, _ := strconv.Atoi(powerStr)

	list, err := h.repo.GetPublishedDevices(minWatts)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var dtos []DeviceViewDTO
	for i := range list {
		dtos = append(dtos, h.toDTO(&list[i]))
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"Devices":        dtos,
		"MinPowerFilter": powerStr,
	})
}

func (h *Handler) PostCreateDraft(c *gin.Context) {
	userID := uint(1)
	existing, _ := h.repo.GetUserDraft(userID)
	if existing != nil {
		c.Redirect(http.StatusFound, "/office-devices/draft")
		return
	}

	name := c.PostForm("device_name")
	newDev := ds.OfficeDevice{
		DeviceName: name,
		CreatorID:  userID,
	}

	_ = h.repo.CreateDraft(&newDev)
	c.Redirect(http.StatusFound, "/office-devices/draft")
}

func (h *Handler) PostPublishDraft(c *gin.Context) {
	userID := uint(1)
	desc := c.PostForm("device_description")
	watts, _ := strconv.Atoi(c.PostForm("active_power_watts"))
	pf, _ := strconv.ParseFloat(c.PostForm("power_factor"), 64)

	_ = h.repo.PublishDraft(userID, desc, watts, pf)
	c.Redirect(http.StatusFound, "/office-devices/grid")
}

func (h *Handler) PostDeleteDevice(c *gin.Context) {
	idStr := c.PostForm("device_id")
	id, _ := strconv.Atoi(idStr)

	_ = h.repo.LogicalDeleteSQL(uint(id))
	c.Redirect(http.StatusFound, "/office-devices/grid")
}