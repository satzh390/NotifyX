package utils

import (
	"strconv"
	"strings"

	"github.com/notifyx/core/domain"
)

// QueryParser provides methods to parse query parameters
type QueryParser interface {
	Query(key string) string
}

// ParseListOptions parses query parameters for pagination, sorting, and filtering
func ParseListOptions(parser QueryParser, orgID string) domain.ListOptions {
	opts := domain.ListOptions{
		OrgID: orgID,
	}

	// Parse pagination
	page := 1
	if pageStr := parser.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := parser.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	opts.Pagination = domain.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	// Parse sorting
	sortByStr := parser.Query("sortBy")
	if sortByStr != "" {
		opts.SortBy = parseSortBy(sortByStr)
	}

	return opts
}

// ParseListOptionsWithGroup parses query parameters including group filter
func ParseListOptionsWithGroup(parser QueryParser, orgID string) domain.ListOptions {
	opts := ParseListOptions(parser, orgID)
	opts.GroupID = parser.Query("groupId")
	return opts
}

// parseSortBy parses sort query parameter
// Format: "field1:asc,field2:desc" or "field1" (defaults to asc)
func parseSortBy(sortByStr string) []domain.SortOption {
	if sortByStr == "" {
		return nil
	}

	var sortOptions []domain.SortOption
	parts := strings.Split(sortByStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon to get field and order
		fieldParts := strings.Split(part, ":")
		field := strings.TrimSpace(fieldParts[0])
		if field == "" {
			continue
		}

		order := domain.SortAsc // default to ascending
		if len(fieldParts) > 1 {
			orderStr := strings.ToLower(strings.TrimSpace(fieldParts[1]))
			if orderStr == "desc" || orderStr == "descending" {
				order = domain.SortDesc
			}
		}

		sortOptions = append(sortOptions, domain.SortOption{
			Field: field,
			Order: order,
		})
	}

	return sortOptions
}

