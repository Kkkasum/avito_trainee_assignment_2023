package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"avito_2023/internal/database"
	"avito_2023/internal/segment/repo"
)

type Handler struct {
	repo repo.Repo
}

// @Summary Add Segment
// @Tags segment
// @Description Add new segment with specified slug
// @Accept json
// @Produce json
// @Param body body AddSegmentRequest true "segment slug"
// @Success 201
// @Failure 400
// @Failure 500
// @Router /segment/add [post]
func (h *Handler) addSegment(c *gin.Context) {
	var body AddSegmentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Percentage < 0 || body.Percentage > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid percentage"})
		return
	}

	if err := h.repo.AddSegment(c.Request.Context(), body.Slug, body.Percentage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "new segment added"})
}

// @Summary Delete Segment
// @Tags segment
// @Description Delete segment with specified slug
// @Accept json
// @Produce json
// @Param body body DeleteSegmentRequest true "segment slug"
// @Success 204
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /segment/delete [delete]
func (h *Handler) deleteSegment(c *gin.Context) {
	var body DeleteSegmentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.DeleteSegment(c.Request.Context(), body.Slug); err != nil {
		if database.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("segment %s not found", body.Slug)})
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
	router := r.Group("segment")

	{
		router.POST("add", h.addSegment)
		router.DELETE("delete", h.deleteSegment)
	}
}
