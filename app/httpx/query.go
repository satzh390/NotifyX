package httpx

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/core/domain"
)

// ParseListOptions parses query parameters for pagination, sorting, and filtering
func ParseListOptions(parser *fiber.Ctx, customerID string) domain.ListOptions {
	opts := domain.ListOptions{
		Filter: map[string]string{"customerId": customerID},
	}

	// Parse pagination (0-based)
	page := 0
	if pageStr := parser.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
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

// parseSortBy parses sort query parameter
// Format: "field1:asc,field2:desc" or "field1" (defaults to asc)
func parseSortBy(sortByStr string) []domain.SortParams {
	if sortByStr == "" {
		return nil
	}

	var sortOptions []domain.SortParams
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

		sortOptions = append(sortOptions, domain.SortParams{
			Field: field,
			Order: order,
		})
	}

	return sortOptions
}
