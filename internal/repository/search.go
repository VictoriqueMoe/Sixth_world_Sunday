package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type (
	SearchEntityType string

	SearchResult struct {
		EntityType        SearchEntityType
		ID                string
		ParentID          *string
		ParentTitle       *string
		Title             string
		Snippet           string
		AuthorID          *string
		AuthorUsername    string
		AuthorDisplayName string
		AuthorAvatarURL   string
		CreatedAt         string
		Rank              float64
	}

	SearchSource struct {
		Type SearchEntityType

		From         string
		AuthorJoin   string
		ParentJoin   string
		IDExpr       string
		TitleExpr    string
		BodyExpr     string
		SearchVector string
		CreatedAt    string

		ParentIDExpr    string
		ParentTitleExpr string

		TrigramOnTitle bool
		TrigramExprs   []string

		ExtraWhere string
	}

	SearchRepository interface {
		Search(ctx context.Context, query string, types []SearchEntityType, limit, offset int) ([]SearchResult, int, error)
		QuickSearch(ctx context.Context, query string, perTypeLimit int) ([]SearchResult, error)
	}

	searchRepository struct {
		db *sql.DB
	}
)

const (
	SearchEntityUser        SearchEntityType = "user"
	SearchEntityChatMessage SearchEntityType = "chat_message"
)

const searchHeadlineOptions = `'MaxFragments=1, MaxWords=18, MinWords=5, ShortWord=3, HighlightAll=false, StartSel=<mark>, StopSel=</mark>'`

var searchSources = []SearchSource{
	{
		Type:           SearchEntityUser,
		From:           "users u",
		IDExpr:         "u.id::text",
		TitleExpr:      "u.display_name",
		BodyExpr:       "u.bio",
		SearchVector:   "u.search_vector",
		CreatedAt:      "u.created_at",
		TrigramOnTitle: true,
		TrigramExprs:   []string{"u.username"},
	},
}

var searchSourcesByType = func() map[SearchEntityType]SearchSource {
	m := make(map[SearchEntityType]SearchSource, len(searchSources))
	for _, s := range searchSources {
		m[s.Type] = s
	}
	return m
}()

func SearchSourceFor(t SearchEntityType) (SearchSource, bool) {
	src, ok := searchSourcesByType[t]
	return src, ok
}

func SearchSources() []SearchSource {
	return searchSources
}

func resolveSearchTypes(types []SearchEntityType) []SearchSource {
	if len(types) == 0 {
		return searchSources
	}
	out := make([]SearchSource, 0, len(types))
	seen := make(map[SearchEntityType]bool, len(types))
	for _, t := range types {
		if seen[t] {
			continue
		}
		src, ok := searchSourcesByType[t]
		if !ok {
			continue
		}
		seen[t] = true
		out = append(out, src)
	}
	return out
}

func (s SearchSource) buildSubquery() string {
	parentIDExpr := s.ParentIDExpr
	if parentIDExpr == "" {
		parentIDExpr = "NULL::text"
	}
	parentTitleExpr := s.ParentTitleExpr
	if parentTitleExpr == "" {
		parentTitleExpr = "NULL::text"
	}

	rankExpr := "ts_rank_cd(" + s.SearchVector + ", q.tsq)"
	matchExpr := s.SearchVector + " @@ q.tsq"

	trigramExprs := append([]string(nil), s.TrigramExprs...)
	if s.TrigramOnTitle {
		trigramExprs = append([]string{s.TitleExpr}, trigramExprs...)
	}
	for _, expr := range trigramExprs {
		rankExpr += " + COALESCE(similarity(" + expr + ", q.qstr), 0)"
		matchExpr += " OR " + expr + " % q.qstr"
	}
	if len(trigramExprs) > 0 {
		matchExpr = "(" + matchExpr + ")"
	}

	var parts []string
	parts = append(parts, "FROM "+s.From)
	if s.ParentJoin != "" {
		parts = append(parts, s.ParentJoin)
	}
	if s.AuthorJoin != "" {
		parts = append(parts, s.AuthorJoin)
	}
	parts = append(parts, "CROSS JOIN q")

	whereParts := []string{"u.banned_at IS NULL", "u.locked_at IS NULL"}
	if s.ExtraWhere != "" {
		whereParts = append(whereParts, s.ExtraWhere)
	}
	whereParts = append(whereParts, matchExpr)

	return fmt.Sprintf(`SELECT '%s' AS entity_type, %s AS id, %s AS parent_id, %s AS parent_title,
            %s AS title,
            ts_headline('english', %s, q.tsq, %s) AS snippet,
            u.id::text AS author_id, u.username AS author_username, u.display_name AS author_display_name, u.avatar_url AS author_avatar_url,
            %s AS created_at,
            (%s)::float8 AS rank
        %s
        WHERE %s`,
		s.Type, s.IDExpr, parentIDExpr, parentTitleExpr,
		s.TitleExpr,
		s.BodyExpr, searchHeadlineOptions,
		s.CreatedAt,
		rankExpr,
		strings.Join(parts, "\n        "),
		strings.Join(whereParts, "\n          AND "),
	)
}

func (r *searchRepository) Search(ctx context.Context, query string, types []SearchEntityType, limit, offset int) ([]SearchResult, int, error) {
	srcs := resolveSearchTypes(types)
	if len(srcs) == 0 {
		return nil, 0, nil
	}

	subqueries := make([]string, len(srcs))
	for i, src := range srcs {
		subqueries[i] = src.buildSubquery()
	}
	union := strings.Join(subqueries, "\nUNION ALL\n")

	countSQL := fmt.Sprintf(`WITH q AS (SELECT websearch_to_tsquery('english', $1) AS tsq, $1 AS qstr)
        SELECT COUNT(*) FROM (%s) results`, union)

	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("search count: %w", err)
	}

	dataSQL := fmt.Sprintf(`WITH q AS (SELECT websearch_to_tsquery('english', $1) AS tsq, $1 AS qstr)
        SELECT entity_type, id, parent_id, parent_title, title, snippet,
               author_id, author_username, author_display_name, author_avatar_url, created_at, rank
        FROM (%s) results
        ORDER BY rank DESC, created_at DESC
        LIMIT $2 OFFSET $3`, union)

	rows, err := r.db.QueryContext(ctx, dataSQL, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	results, err := scanSearchRows(rows, limit)
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *searchRepository) QuickSearch(ctx context.Context, query string, perTypeLimit int) ([]SearchResult, error) {
	subqueries := make([]string, len(searchSources))
	for i, src := range searchSources {
		subqueries[i] = fmt.Sprintf(`(SELECT * FROM (%s) sub ORDER BY rank DESC, created_at DESC LIMIT %d)`, src.buildSubquery(), perTypeLimit)
	}
	union := strings.Join(subqueries, "\nUNION ALL\n")

	sqlStr := fmt.Sprintf(`WITH q AS (SELECT websearch_to_tsquery('english', $1) AS tsq, $1 AS qstr)
        SELECT entity_type, id, parent_id, parent_title, title, snippet,
               author_id, author_username, author_display_name, author_avatar_url, created_at, rank
        FROM (%s) results
        ORDER BY rank DESC, created_at DESC`, union)

	rows, err := r.db.QueryContext(ctx, sqlStr, query)
	if err != nil {
		return nil, fmt.Errorf("quick search: %w", err)
	}
	defer rows.Close()

	return scanSearchRows(rows, perTypeLimit*len(searchSources))
}

func scanSearchRows(rows *sql.Rows, capacity int) ([]SearchResult, error) {
	results := make([]SearchResult, 0, capacity)
	for rows.Next() {
		var (
			r         SearchResult
			createdAt time.Time
			entityT   string
		)
		if err := rows.Scan(
			&entityT, &r.ID, &r.ParentID, &r.ParentTitle, &r.Title, &r.Snippet,
			&r.AuthorID, &r.AuthorUsername, &r.AuthorDisplayName, &r.AuthorAvatarURL,
			&createdAt, &r.Rank,
		); err != nil {
			return nil, fmt.Errorf("search scan: %w", err)
		}
		r.EntityType = SearchEntityType(entityT)
		r.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search rows: %w", err)
	}
	return results, nil
}
