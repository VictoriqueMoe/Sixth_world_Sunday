package dto

type (
	SaveMapRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		SourceURL   string `json:"source_url"`
	}

	MapResponse struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		EmbedURL    string `json:"embed_url"`
		SourceURL   string `json:"source_url"`
		CanManage   bool   `json:"can_manage"`
		CreatedBy   string `json:"created_by"`
		CreatedAt   string `json:"created_at"`
	}

	MapListResponse struct {
		Maps []MapResponse `json:"maps"`
	}
)
