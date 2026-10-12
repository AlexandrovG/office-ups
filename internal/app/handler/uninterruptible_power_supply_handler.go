package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"office-uninterruptible_power_supply/internal/app/ds"
	"office-uninterruptible_power_supply/internal/app/repository"
	"office-uninterruptible_power_supply/internal/app/singleton"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{repo: r}
}

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
		result = append(result, toUninterruptiblePowerSupplySerializer(&list[i]))
	}
	c.JSON(http.StatusOK, result)
}

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
	c.JSON(http.StatusOK, toUninterruptiblePowerSupplySerializer(ups))
}

func (h *Handler) GetUninterruptiblePowerSupplyDraftAPI(c *gin.Context) {
	draft, err := h.repo.GetUserDraft(singleton.CurrentUser())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if draft == nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, toUninterruptiblePowerSupplySerializer(draft))
}

func (h *Handler) CreateUninterruptiblePowerSupplyAPI(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	existing, _ := h.repo.GetUserDraft(singleton.CurrentUser())
	if existing != nil {
		c.Status(http.StatusConflict)
		return
	}

	ups := ds.UninterruptiblePowerSupply{
		UpsName:   c.PostForm("ups_name"),
		CreatorID: singleton.CurrentUser(),
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
	c.JSON(http.StatusCreated, toUninterruptiblePowerSupplySerializer(updated))
}

func (h *Handler) PublishUninterruptiblePowerSupplyAPI(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	draft, err := h.repo.GetUserDraft(singleton.CurrentUser())
	if err != nil || draft == nil {
		c.Status(http.StatusNotFound)
		return
	}

	desc := c.PostForm("description")
	capacity, _ := strconv.ParseFloat(c.PostForm("capacity"), 64)
	maxCurrent, _ := strconv.ParseFloat(c.PostForm("max_current"), 64)

	if err := h.repo.PublishUninterruptiblePowerSupplyDraft(singleton.CurrentUser(), draft.ID, desc, capacity, maxCurrent); err != nil {
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
	c.JSON(http.StatusOK, toUninterruptiblePowerSupplySerializer(updated))
}

func (h *Handler) DeleteUninterruptiblePowerSupplyAPI(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.repo.LogicalDeleteUninterruptiblePowerSupplySQL(uint(id), singleton.CurrentUser()); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) LikeUninterruptiblePowerSupplyAPI(c *gin.Context) {
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

	if err := h.repo.LikeUninterruptiblePowerSupply(singleton.CurrentUser(), uint(id), *req.Value); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RegisterAPI(c *gin.Context) {
	var req ds.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := h.repo.CreateUser(req.Login, req.Password)
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	c.JSON(http.StatusCreated, ds.UserSerializer{
		ID:          user.ID,
		Login:       user.UserLogin,
		IsModerator: user.IsModerator,
	})
}

func (h *Handler) LoginAPI(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (h *Handler) LogoutAPI(c *gin.Context) {
	c.Status(http.StatusOK)
}

func toUninterruptiblePowerSupplySerializer(u *ds.UninterruptiblePowerSupply) ds.UninterruptiblePowerSupplySerializer {
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
	}
}