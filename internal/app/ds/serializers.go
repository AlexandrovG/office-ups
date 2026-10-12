package ds

type UninterruptiblePowerSupplySerializer struct {
	ID          uint    `json:"id"`
	UpsName     string  `json:"ups_name"`
	Description string  `json:"description"`
	Capacity    float64 `json:"capacity"`
	MaxCurrent  float64 `json:"max_current"`
	ImageURL    string  `json:"image_url"`
	VideoURL    string  `json:"video_url"`
	LikesCount  int     `json:"likes_count"`
}

type UninterruptiblePowerSupplyLikeRequest struct {
	Value *int `json:"value" binding:"required,oneof=0 1"`
}

type UninterruptiblePowerSupplyPublishRequest struct {
	Description string  `json:"description"`
	Capacity    float64 `json:"capacity"`
	MaxCurrent  float64 `json:"max_current"`
}

type RegisterRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserSerializer struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}