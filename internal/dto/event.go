package dto

type (
	EventFields struct {
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		CoverURL     string  `json:"cover_url"`
		LocationType string  `json:"location_type"`
		VoiceRoomID  *string `json:"voice_room_id"`
		ExternalURL  string  `json:"external_url"`
		StartAt      string  `json:"start_at"`
		Frequency    string  `json:"frequency"`
	}

	CreateEventRequest struct {
		EventFields
	}

	UpdateEventRequest struct {
		EventFields
	}

	EventResponse struct {
		ID               string   `json:"id"`
		Title            string   `json:"title"`
		Description      string   `json:"description"`
		CoverURL         string   `json:"cover_url"`
		LocationType     string   `json:"location_type"`
		VoiceRoomID      *string  `json:"voice_room_id"`
		VoiceRoomName    string   `json:"voice_room_name"`
		ExternalURL      string   `json:"external_url"`
		StartAt          string   `json:"start_at"`
		Frequency        string   `json:"frequency"`
		NextStartAt      string   `json:"next_start_at"`
		NextOccurrences  []string `json:"next_occurrences"`
		RSVPCount        int      `json:"rsvp_count"`
		ViewerInterested bool     `json:"viewer_interested"`
		RSVPAvatars      []string `json:"rsvp_avatars"`
		CanManage        bool     `json:"can_manage"`
		CreatedBy        string   `json:"created_by"`
		CreatedAt        string   `json:"created_at"`
	}

	EventListResponse struct {
		Events []EventResponse `json:"events"`
	}
)
