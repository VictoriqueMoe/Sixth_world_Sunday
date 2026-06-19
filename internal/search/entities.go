package search

import "Sixth_world_Suday/internal/repository"

func AllEntityTypes() []repository.SearchEntityType {
	srcs := repository.SearchSources()
	out := make([]repository.SearchEntityType, len(srcs))
	for i, s := range srcs {
		out[i] = s.Type
	}
	return out
}

func ChildEntityTypes() []repository.SearchEntityType {
	srcs := repository.SearchSources()
	out := make([]repository.SearchEntityType, 0, len(srcs))
	for _, s := range srcs {
		if s.ParentIDExpr != "" {
			out = append(out, s.Type)
		}
	}
	return out
}
