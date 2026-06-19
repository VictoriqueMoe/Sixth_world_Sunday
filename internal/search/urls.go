package search

import (
	"fmt"
	"net/url"

	"Sixth_world_Suday/internal/repository"
)

func init() {
	for _, t := range AllEntityTypes() {
		if _, ok := urlBuilders[t]; !ok {
			panic(fmt.Sprintf("search.urlBuilders: missing URL builder for entity type %q (registered in repository.searchSources but not in internal/search/urls.go)", t))
		}
	}
}

var urlBuilders = map[repository.SearchEntityType]func(repository.SearchResult) string{
	repository.SearchEntityUser: func(r repository.SearchResult) string {
		return "/user/" + r.AuthorUsername
	},
	repository.SearchEntityChatMessage: func(r repository.SearchResult) string {
		if r.ParentID == nil {
			return ""
		}
		u := "/rooms/" + *r.ParentID
		if r.CreatedAt != "" {
			u += "?at=" + url.QueryEscape(r.CreatedAt)
		}
		return u + "#msg-" + r.ID
	},
}

func BuildURL(r repository.SearchResult) string {
	if fn, ok := urlBuilders[r.EntityType]; ok {
		return fn(r)
	}
	return ""
}
