package handler

type AddSegmentRequest struct {
	Slug       string `json:"slug" binding:"required"`
	Percentage uint   `json:"percentage"`
}

type DeleteSegmentRequest struct {
	Slug string `json:"slug" binding:"required"`
}
