package dto

import "github.com/google/uuid"

type (
	PostMediaResponse struct {
		ID           int    `json:"id"`
		MediaURL     string `json:"media_url"`
		MediaType    string `json:"media_type"`
		ThumbnailURL string `json:"thumbnail_url,omitempty"`
		SortOrder    int    `json:"sort_order"`
	}

	EmbedResponse struct {
		URL      string `json:"url"`
		Type     string `json:"type"`
		Title    string `json:"title,omitempty"`
		Desc     string `json:"description,omitempty"`
		Image    string `json:"image,omitempty"`
		SiteName string `json:"site_name,omitempty"`
		VideoID  string `json:"video_id,omitempty"`
	}

	CreateCommentRequest struct {
		ParentID *uuid.UUID `json:"parent_id,omitempty"`
		Body     string     `json:"body"`
	}

	UpdateCommentRequest struct {
		Body string `json:"body"`
	}
)
