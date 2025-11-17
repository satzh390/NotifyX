package utils

import (
	"testing"

	"github.com/notifyx/core/domain"
)

type mockQueryParser struct {
	queries map[string]string
}

func (m *mockQueryParser) Query(key string) string {
	return m.queries[key]
}

func TestParseListOptions(t *testing.T) {
	tests := []struct {
		name     string
		queries  map[string]string
		orgID    string
		expected domain.ListOptions
	}{
		{
			name:    "default values",
			queries: map[string]string{},
			orgID:   "org-123",
			expected: domain.ListOptions{
				OrgID: "org-123",
				Pagination: domain.PaginationParams{
					Page:     1,
					PageSize: 20,
				},
				SortBy: nil,
			},
		},
		{
			name: "custom pagination",
			queries: map[string]string{
				"page":     "2",
				"pageSize": "50",
			},
			orgID: "org-123",
			expected: domain.ListOptions{
				OrgID: "org-123",
				Pagination: domain.PaginationParams{
					Page:     2,
					PageSize: 50,
				},
				SortBy: nil,
			},
		},
		{
			name: "with sorting",
			queries: map[string]string{
				"page":   "1",
				"sortBy": "createdAt:desc",
			},
			orgID: "org-123",
			expected: domain.ListOptions{
				OrgID: "org-123",
				Pagination: domain.PaginationParams{
					Page:     1,
					PageSize: 20,
				},
				SortBy: []domain.SortOption{
					{Field: "createdAt", Order: domain.SortDesc},
				},
			},
		},
		{
			name: "multiple sort fields",
			queries: map[string]string{
				"sortBy": "name:asc,createdAt:desc",
			},
			orgID: "org-123",
			expected: domain.ListOptions{
				OrgID: "org-123",
				Pagination: domain.PaginationParams{
					Page:     1,
					PageSize: 20,
				},
				SortBy: []domain.SortOption{
					{Field: "name", Order: domain.SortAsc},
					{Field: "createdAt", Order: domain.SortDesc},
				},
			},
		},
		{
			name: "invalid page defaults to 1",
			queries: map[string]string{
				"page": "invalid",
			},
			orgID: "org-123",
			expected: domain.ListOptions{
				OrgID: "org-123",
				Pagination: domain.PaginationParams{
					Page:     1,
					PageSize: 20,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &mockQueryParser{queries: tt.queries}
			result := ParseListOptions(parser, tt.orgID)

			if result.OrgID != tt.expected.OrgID {
				t.Errorf("OrgID = %v, want %v", result.OrgID, tt.expected.OrgID)
			}
			if result.Pagination.Page != tt.expected.Pagination.Page {
				t.Errorf("Page = %v, want %v", result.Pagination.Page, tt.expected.Pagination.Page)
			}
			if result.Pagination.PageSize != tt.expected.Pagination.PageSize {
				t.Errorf("PageSize = %v, want %v", result.Pagination.PageSize, tt.expected.Pagination.PageSize)
			}
			if len(result.SortBy) != len(tt.expected.SortBy) {
				t.Errorf("SortBy length = %v, want %v", len(result.SortBy), len(tt.expected.SortBy))
			}
			for i, sort := range result.SortBy {
				if i >= len(tt.expected.SortBy) {
					break
				}
				if sort.Field != tt.expected.SortBy[i].Field {
					t.Errorf("SortBy[%d].Field = %v, want %v", i, sort.Field, tt.expected.SortBy[i].Field)
				}
				if sort.Order != tt.expected.SortBy[i].Order {
					t.Errorf("SortBy[%d].Order = %v, want %v", i, sort.Order, tt.expected.SortBy[i].Order)
				}
			}
		})
	}
}

func TestParseListOptionsWithGroup(t *testing.T) {
	parser := &mockQueryParser{
		queries: map[string]string{
			"groupId": "group-123",
			"page":    "1",
		},
	}
	result := ParseListOptionsWithGroup(parser, "org-123")

	if result.OrgID != "org-123" {
		t.Errorf("OrgID = %v, want org-123", result.OrgID)
	}
	if result.GroupID != "group-123" {
		t.Errorf("GroupID = %v, want group-123", result.GroupID)
	}
}

