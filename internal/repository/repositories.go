package repository

import "database/sql"

type (
	Repositories struct {
		db                *sql.DB
		Session           SessionRepository
		User              UserRepository
		Notification      NotificationRepository
		Role              RoleRepository
		Settings          SettingsRepository
		AuditLog          AuditLogRepository
		Stats             StatsRepository
		Invite            InviteRepository
		PasswordReset     PasswordResetRepository
		EmailVerification EmailVerificationRepository
		Chat              ChatRepository
		Report            ReportRepository
		Upload            UploadRepository
		Block             BlockRepository
		VanityRole        VanityRoleRepository
		ChatRoomBan       ChatRoomBanRepository
		ChatBannedWord    ChatBannedWordRepository
		Search            SearchRepository
	}
)

func (r *Repositories) DB() *sql.DB {
	return r.db
}

func New(db *sql.DB) *Repositories {
	return &Repositories{
		db:                db,
		Session:           &sessionRepository{db: db},
		User:              &userRepository{db: db},
		Notification:      &notificationRepository{db: db},
		Role:              &roleRepository{db: db},
		Settings:          &settingsRepository{db: db},
		AuditLog:          &auditLogRepository{db: db},
		Stats:             &statsRepository{db: db},
		Invite:            &inviteRepository{db: db},
		PasswordReset:     &passwordResetRepository{db: db},
		EmailVerification: &emailVerificationRepository{db: db},
		Chat:              &chatRepository{db: db},
		Report:            &reportRepository{db: db},
		Upload:            &uploadRepository{db: db},
		Block:             &blockRepository{db: db},
		VanityRole:        &vanityRoleRepository{db: db},
		ChatRoomBan:       &chatRoomBanRepository{db: db},
		ChatBannedWord:    &chatBannedWordRepository{db: db},
		Search:            &searchRepository{db: db},
	}
}
