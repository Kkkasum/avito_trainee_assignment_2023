package handler

type GetUserSegmentsUri struct {
	UserID uint `uri:"user_id" binding:"required"`
}

type GetUserHistoryUri struct {
	UserID uint `uri:"user_id" binding:"required"`
}

type GetUserHistoryQuery struct {
	Month uint `form:"month" binding:"required"`
	Year  uint `form:"year" binding:"required"`
}

type UpdateUserSegmentsRequest struct {
	UserID     uint     `json:"user_id" binding:"required"`
	SlugsToAdd []string `json:"slugs_to_add" binding:"required"`
	SlugsToDel []string `json:"slugs_to_del" binding:"required"`
	DeleteAt   int64    `json:"delete_at"`
}
