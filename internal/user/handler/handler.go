package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"avito_2023/internal/database"
	"avito_2023/internal/user/repo"
)

type Handler struct {
	repo repo.Repo
}

// @Summary Get User Segments
// @Tags user
// @Description Get active segments for specified user
// @Accept json
// @Produce json
// @Param user_id path int true "user ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /user/{user_id} [get]
func (h *Handler) getUserSegments(c *gin.Context) {
	var uri GetUserSegmentsUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	segments, err := h.repo.GetUserSegments(c.Request.Context(), uri.UserID)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("segments for user %d not found", uri.UserID)})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": uri.UserID, "segments": segments})
}

// @Summary Get User History
// @Tags user
// @Description Get user segments history
// @Accept json
// @Produce json
// @Param user_id path int true "user ID"
// @Param month query int true "month"
// @Param year query int true "year"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /user/history/{user_id} [get]
func (h *Handler) getUserHistory(c *gin.Context) {
	var uri GetUserHistoryUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query GetUserHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history, err := h.repo.GetUserHistory(c.Request.Context(), uri.UserID, query.Month, query.Year)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("history for user %d not found", uri.UserID)})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": uri.UserID, "history": history})
}

// @Summary Update User Segments
// @Tags user
// @Description Update user segments with specified slugs for specified user
// @Accept json
// @Produce json
// @Param body body UpdateUserSegmentsRequest true "user and segments info"
// @Success 204
// @Failure 400
// @Failure 500
// @Router /user/segment [put]
func (h *Handler) updateUserSegments(c *gin.Context) {
	var body UpdateUserSegmentsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.DeleteAt != 0 && body.DeleteAt < time.Now().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value delete_at"})
		return
	}

	var deleteAt *time.Time
	if body.DeleteAt != 0 {
		tmp := time.Unix(body.DeleteAt, 0)
		deleteAt = &tmp
	}

	if err := h.repo.UpdateUserSegments(c.Request.Context(), body.UserID, body.SlugsToAdd, body.SlugsToDel, deleteAt); err != nil {
		if database.IsUpdateUserSegmentsInvalidSegmentsErr(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segments"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func NewHandler(repo repo.Repo) *Handler {
	return &Handler{
		repo: repo,
	}
}

func Route(r *gin.Engine, h *Handler) {
	router := r.Group("user")

	{
		router.GET("/:user_id", h.getUserSegments)
		router.GET("/history/:user_id", h.getUserHistory)
		router.PUT("/segment", h.updateUserSegments)
	}
}
