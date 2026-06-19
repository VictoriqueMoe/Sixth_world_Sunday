package dto

import (
	"Sixth_world_Suday/internal/role"

	"github.com/google/uuid"
)

type (
	UserResponse struct {
		ID          uuid.UUID            `json:"id"`
		Username    string               `json:"username"`
		DisplayName string               `json:"display_name"`
		AvatarURL   string               `json:"avatar_url,omitempty"`
		Role        role.Role            `json:"role,omitempty"`
		VanityRoles []VanityRoleResponse `json:"vanity_roles,omitempty"`
		Banned      bool                 `json:"banned,omitempty"`
		BanReason   string               `json:"ban_reason,omitempty"`
		Locked      bool                 `json:"locked,omitempty"`
		LockReason  string               `json:"lock_reason,omitempty"`
	}

	UserProfileResponse struct {
		UserResponse
		Bio                string  `json:"bio"`
		BannerURL          string  `json:"banner_url"`
		BannerPosition     float64 `json:"banner_position"`
		FavouriteCharacter string  `json:"favourite_character"`
		Gender             string  `json:"gender"`
		PronounSubject     string  `json:"pronoun_subject"`
		PronounPossessive  string  `json:"pronoun_possessive"`
		Online             bool    `json:"online"`
		SocialTwitter      string  `json:"social_twitter"`
		SocialDiscord      string  `json:"social_discord"`
		SocialWaifulist    string  `json:"social_waifulist"`
		SocialTumblr       string  `json:"social_tumblr"`
		SocialGithub       string  `json:"social_github"`
		Website            string  `json:"website"`
		DmsEnabled         bool    `json:"dms_enabled"`
		DOB                string  `json:"dob,omitempty"`
		DOBPublic          bool    `json:"dob_public"`
		Email              string  `json:"email,omitempty"`
		EmailPublic        bool    `json:"email_public"`
		CreatedAt          string  `json:"created_at"`
		// Only present when viewing own profile
		Private *UserPrivateFields `json:"private,omitempty"`
	}

	UserPrivateFields struct {
		EmailVerified         bool   `json:"email_verified"`
		VerifyGraceUntil      string `json:"verify_grace_until,omitempty"`
		EmailNotifications    bool   `json:"email_notifications"`
		PlayMessageSound      bool   `json:"play_message_sound"`
		PlayNotificationSound bool   `json:"play_notification_sound"`
		DefaultProfileTab     string `json:"default_profile_tab"`
		Theme                 string `json:"theme"`
		Font                  string `json:"font"`
		WideLayout            bool   `json:"wide_layout"`
	}

	UpdateProfileRequest struct {
		DisplayName           string  `json:"display_name"`
		Bio                   string  `json:"bio"`
		AvatarURL             string  `json:"avatar_url"`
		BannerURL             string  `json:"banner_url"`
		BannerPosition        float64 `json:"banner_position"`
		FavouriteCharacter    string  `json:"favourite_character"`
		Gender                string  `json:"gender"`
		PronounSubject        string  `json:"pronoun_subject"`
		PronounPossessive     string  `json:"pronoun_possessive"`
		SocialTwitter         string  `json:"social_twitter"`
		SocialDiscord         string  `json:"social_discord"`
		SocialWaifulist       string  `json:"social_waifulist"`
		SocialTumblr          string  `json:"social_tumblr"`
		SocialGithub          string  `json:"social_github"`
		Website               string  `json:"website"`
		DmsEnabled            bool    `json:"dms_enabled"`
		DOB                   string  `json:"dob"`
		DOBPublic             bool    `json:"dob_public"`
		Email                 string  `json:"email"`
		EmailPublic           bool    `json:"email_public"`
		EmailNotifications    bool    `json:"email_notifications"`
		PlayMessageSound      bool    `json:"play_message_sound"`
		PlayNotificationSound bool    `json:"play_notification_sound"`
		DefaultProfileTab     string  `json:"default_profile_tab"`
	}

	ChangePasswordRequest struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	ForgotPasswordRequest struct {
		Username       string `json:"username"`
		TurnstileToken string `json:"turnstile_token,omitempty"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	DeleteAccountRequest struct {
		Password string `json:"password"`
	}

	Credentials interface {
		GetUsername() string
		GetPassword() string
	}

	LoginRequest struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		TurnstileToken string `json:"turnstile_token,omitempty"`
	}

	RegisterRequest struct {
		LoginRequest
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		InviteCode  string `json:"invite_code,omitempty"`
	}

	SetEmailRequest struct {
		Email string `json:"email"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token"`
	}
)

func (r LoginRequest) GetUsername() string { return r.Username }
func (r LoginRequest) GetPassword() string { return r.Password }
