package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"office-uninterruptible_power_supply/internal/app/ds"
	"office-uninterruptible_power_supply/internal/app/redis"
	"office-uninterruptible_power_supply/internal/app/repository"
)

type Handler struct {
	repo  *repository.Repository
	redis *redis.Client
}

func NewHandler(r *repository.Repository, rc *redis.Client) *Handler {
	return &Handler{repo: r, redis: rc}
}

// GetUninterruptiblePowerSuppliesListAPI godoc
// @Summary      Список ИБП с фильтром
// @Description  Возвращает опубликованные ИБП, можно фильтровать по ёмкости
// @Tags         uninterruptible-power-supplies
// @Produce      json
// @Param        min_capacity  query  number  false  "Минимальная ёмкость"
// @Param        max_capacity  query  number  false  "Максимальная ёмкость"
// @Success      200  {array}   ds.UninterruptiblePowerSupplySerializer
// @Router       /uninterruptible-power-supplies [get]
func (h *Handler) GetUninterruptiblePowerSuppliesListAPI(c *gin.Context) {
	minCap, _ := strconv.ParseFloat(c.Query("min_capacity"), 64)
	maxCap, _ := strconv.ParseFloat(c.Query("max_capacity"), 64)

	list, err := h.repo.GetPublishedUninterruptiblePowerSupplies(minCap, maxCap)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	result := make([]ds.UninterruptiblePowerSupplySerializer, 0, len(list))
	for i := range list {
		result = append(result, toSerializer(&list[i], currentUserFromContext(c)))
	}
	c.JSON(http.StatusOK, result)
}

// GetUninterruptiblePowerSuppliesReelsAPI godoc
// @Summary      Лента ИБП
// @Tags         uninterruptible-power-supplies
// @Produce      json
// @Param        id    query  int   false  "ID текущей услуги"
// @Param        next  query  bool  false  "Перейти к следующей"
// @Success      200  {object}  ds.UninterruptiblePowerSupplySerializer
// @Failure      404
// @Router       /uninterruptible-power-supplies/reels [get]
func (h *Handler) GetUninterruptiblePowerSuppliesReelsAPI(c *gin.Context) {
	idStr := c.Query("id")
	next := c.Query("next")

	var ups *ds.UninterruptiblePowerSupply
	var err error

	if idStr == "" {
		list, _ := h.repo.GetPublishedUninterruptiblePowerSupplies(0, 0)
		if len(list) > 0 {
			ups = &list[0]
		}
	} else {
		id, _ := strconv.Atoi(idStr)
		if next == "true" {
			ups, _ = h.repo.GetNextPublishedUninterruptiblePowerSupply(uint(id))
		} else {
			ups, err = h.repo.GetUninterruptiblePowerSupplyByID(uint(id))
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
		}
	}

	if ups == nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, toSerializer(ups, currentUserFromContext(c)))
}

// GetUninterruptiblePowerSupplyDraftAPI godoc
// @Summary      Черновик пользователя
// @Tags         uninterruptible-power-supplies
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ds.UninterruptiblePowerSupplySerializer
// @Failure      404
// @Router       /uninterruptible-power-supplies/draft [get]
func (h *Handler) GetUninterruptiblePowerSupplyDraftAPI(c *gin.Context) {
	userID := currentUserFromContext(c)
	draft, err := h.repo.GetUserDraft(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if draft == nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, toSerializer(draft, userID))
}

// CreateUninterruptiblePowerSupplyAPI godoc
// @Summary      Создать черновик ИБП с файлами
// @Tags         uninterruptible-power-supplies
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        ups_name  formData  string  true   "Название"
// @Param        image     formData  file    false  "Картинка"
// @Param        video     formData  file    false  "Видео"
// @Success      201  {object}  ds.UninterruptiblePowerSupplySerializer
// @Failure      409
// @Router       /uninterruptible-power-supplies [post]
func (h *Handler) CreateUninterruptiblePowerSupplyAPI(c *gin.Context) {
	userID := currentUserFromContext(c)

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	existing, _ := h.repo.GetUserDraft(userID)
	if existing != nil {
		c.Status(http.StatusConflict)
		return
	}

	ups := ds.UninterruptiblePowerSupply{
		UpsName:   c.PostForm("ups_name"),
		CreatorID: userID,
	}

	if err := h.repo.CreateUninterruptiblePowerSupplyDraft(&ups); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if header, err := c.FormFile("image"); err == nil {
		h.repo.UploadUninterruptiblePowerSupplyImage(ups.ID, header)
	}
	if header, err := c.FormFile("video"); err == nil {
		h.repo.UploadUninterruptiblePowerSupplyVideo(ups.ID, header)
	}

	updated, _ := h.repo.GetUninterruptiblePowerSupplyByID(ups.ID)
	c.JSON(http.StatusCreated, toSerializer(updated, userID))
}

// PublishUninterruptiblePowerSupplyAPI godoc
// @Summary      Публикация услуги
// @Tags         uninterruptible-power-supplies
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        description   formData  string  false  "Описание"
// @Param        capacity      formData  number  false  "Ёмкость"
// @Param        max_current   formData  number  false  "Ток"
// @Param        image         formData  file    false  "Картинка"
// @Param        video         formData  file    false  "Видео"
// @Success      200  {object}  ds.UninterruptiblePowerSupplySerializer
// @Router       /uninterruptible-power-supplies/publish [put]
func (h *Handler) PublishUninterruptiblePowerSupplyAPI(c *gin.Context) {
	userID := currentUserFromContext(c)

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	draft, err := h.repo.GetUserDraft(userID)
	if err != nil || draft == nil {
		c.Status(http.StatusNotFound)
		return
	}

	desc := c.PostForm("description")
	capacity, _ := strconv.ParseFloat(c.PostForm("capacity"), 64)
	maxCurrent, _ := strconv.ParseFloat(c.PostForm("max_current"), 64)

	if err := h.repo.PublishUninterruptiblePowerSupplyDraft(userID, draft.ID, desc, capacity, maxCurrent); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if header, err := c.FormFile("image"); err == nil {
		h.repo.UploadUninterruptiblePowerSupplyImage(draft.ID, header)
	}
	if header, err := c.FormFile("video"); err == nil {
		h.repo.UploadUninterruptiblePowerSupplyVideo(draft.ID, header)
	}

	updated, _ := h.repo.GetUninterruptiblePowerSupplyByID(draft.ID)
	c.JSON(http.StatusOK, toSerializer(updated, userID))
}

// DeleteUninterruptiblePowerSupplyAPI godoc
// @Summary      Удалить услугу
// @Tags         uninterruptible-power-supplies
// @Security     BearerAuth
// @Param        id  path  int  true  "ID услуги"
// @Success      204
// @Router       /uninterruptible-power-supplies/{id} [delete]
func (h *Handler) DeleteUninterruptiblePowerSupplyAPI(c *gin.Context) {
	userID := currentUserFromContext(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.repo.LogicalDeleteUninterruptiblePowerSupplySQL(uint(id), userID); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

// LikeUninterruptiblePowerSupplyAPI godoc
// @Summary      Лайк/анлайк
// @Tags         uninterruptible-power-supplies
// @Security     BearerAuth
// @Accept       json
// @Param        id     path  int     true  "ID услуги"
// @Param        value  body  object  true  "value: 0 или 1"
// @Success      204
// @Router       /uninterruptible-power-supplies/{id}/like [post]
func (h *Handler) LikeUninterruptiblePowerSupplyAPI(c *gin.Context) {
	userID := currentUserFromContext(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var req ds.UninterruptiblePowerSupplyLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.Value == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.repo.LikeUninterruptiblePowerSupply(userID, uint(id), *req.Value); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

func toSerializer(u *ds.UninterruptiblePowerSupply, currentUserID uint) ds.UninterruptiblePowerSupplySerializer {
	img := ""
	if u.ImageURL.Valid {
		img = u.ImageURL.String
	}
	vid := ""
	if u.VideoURL.Valid {
		vid = u.VideoURL.String
	}

	return ds.UninterruptiblePowerSupplySerializer{
		ID:          u.ID,
		UpsName:     u.UpsName,
		Description: u.Description,
		Capacity:    u.Capacity,
		MaxCurrent:  u.MaxCurrent,
		ImageURL:    img,
		VideoURL:    vid,
		LikesCount:  len(u.Likes),
		IsCreator:   u.CreatorID == currentUserID,
		IsLiked:     hasLike(u.Likes, currentUserID),
	}
}

func hasLike(likes []ds.UninterruptiblePowerSupplyLike, userID uint) bool {
	for _, l := range likes {
		if l.UserID == userID {
			return true
		}
	}
	return false
}