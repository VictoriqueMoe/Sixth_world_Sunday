package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type (
	SiteStats struct {
		TotalUsers     int
		TotalMessages  int
		TotalRooms     int
		NewUsers24h    int
		NewUsers7d     int
		NewUsers30d    int
		NewMessages24h int
		NewMessages7d  int
		NewMessages30d int
	}

	ActiveUser struct {
		ID          uuid.UUID
		Username    string
		DisplayName string
		AvatarURL   string
		ActionCount int
	}

	StatsRepository interface {
		GetOverview(ctx context.Context) (*SiteStats, error)
		GetMostActiveUsers(ctx context.Context, limit int) ([]ActiveUser, error)
	}

	statsRepository struct {
		db *sql.DB
	}
)

func (r *statsRepository) GetOverview(ctx context.Context) (*SiteStats, error) {
	var s SiteStats

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&s.TotalUsers); err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chat_messages`).Scan(&s.TotalMessages)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chat_rooms`).Scan(&s.TotalRooms)

	periods := []struct {
		interval string
		users    *int
		messages *int
	}{
		{"1 day", &s.NewUsers24h, &s.NewMessages24h},
		{"7 days", &s.NewUsers7d, &s.NewMessages7d},
		{"30 days", &s.NewUsers30d, &s.NewMessages30d},
	}

	for _, p := range periods {
		_ = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM users WHERE created_at > NOW() - $1::interval`, p.interval,
		).Scan(p.users)
		_ = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM chat_messages WHERE created_at > NOW() - $1::interval`, p.interval,
		).Scan(p.messages)
	}

	return &s, nil
}

func (r *statsRepository) GetMostActiveUsers(ctx context.Context, limit int) ([]ActiveUser, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT u.id, u.username, u.display_name, u.avatar_url, COUNT(m.id) as action_count
		 FROM chat_messages m
		 JOIN users u ON m.sender_id = u.id
		 GROUP BY u.id
		 ORDER BY action_count DESC
		 LIMIT $1`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("most active users: %w", err)
	}
	defer rows.Close()

	var users []ActiveUser
	for rows.Next() {
		var u ActiveUser
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.ActionCount); err != nil {
			return nil, fmt.Errorf("scan active user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
