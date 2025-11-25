package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestParseListOptions_DefaultPagination(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		opts := ParseListOptions(c, "test-org")
		return c.JSON(opts)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var opts domain.ListOptions
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 0, opts.Pagination.Page, "Default page should be 0")
	assert.Equal(t, 20, opts.Pagination.PageSize, "Default pageSize should be 20")
	assert.Equal(t, "test-org", opts.Filter["customerId"])
}

func TestParseListOptions_CustomPagination(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		opts := ParseListOptions(c, "test-org")
		return c.JSON(opts)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?page=2&pageSize=50", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var opts domain.ListOptions
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 2, opts.Pagination.Page)
	assert.Equal(t, 50, opts.Pagination.PageSize)
}

func TestParseListOptions_InvalidPage(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		opts := ParseListOptions(c, "test-org")
		return c.JSON(opts)
	})

	// Test negative page
	req := httptest.NewRequest(http.MethodGet, "/test?page=-1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	var opts domain.ListOptions
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 0, opts.Pagination.Page, "Negative page should default to 0")

	// Test invalid page (non-numeric)
	req = httptest.NewRequest(http.MethodGet, "/test?page=abc", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 0, opts.Pagination.Page, "Invalid page should default to 0")
}

func TestParseListOptions_InvalidPageSize(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		opts := ParseListOptions(c, "test-org")
		return c.JSON(opts)
	})

	// Test negative pageSize
	req := httptest.NewRequest(http.MethodGet, "/test?pageSize=-1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	var opts domain.ListOptions
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 20, opts.Pagination.PageSize, "Negative pageSize should default to 20")

	// Test invalid pageSize (non-numeric)
	req = httptest.NewRequest(http.MethodGet, "/test?pageSize=abc", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 20, opts.Pagination.PageSize, "Invalid pageSize should default to 20")
}

func TestParseListOptions_Sorting(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		opts := ParseListOptions(c, "test-org")
		return c.JSON(opts)
	})

	// Test single sort field
	req := httptest.NewRequest(http.MethodGet, "/test?sortBy=name:asc", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	var opts domain.ListOptions
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Len(t, opts.SortBy, 1)
	assert.Equal(t, "name", opts.SortBy[0].Field)
	assert.Equal(t, domain.SortAsc, opts.SortBy[0].Order)

	// Test multiple sort fields
	req = httptest.NewRequest(http.MethodGet, "/test?sortBy=name:asc,createdAt:desc", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Len(t, opts.SortBy, 2)
	assert.Equal(t, "name", opts.SortBy[0].Field)
	assert.Equal(t, domain.SortAsc, opts.SortBy[0].Order)
	assert.Equal(t, "createdAt", opts.SortBy[1].Field)
	assert.Equal(t, domain.SortDesc, opts.SortBy[1].Order)

	// Test sort without order (defaults to asc)
	req = httptest.NewRequest(http.MethodGet, "/test?sortBy=name", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Len(t, opts.SortBy, 1)
	assert.Equal(t, "name", opts.SortBy[0].Field)
	assert.Equal(t, domain.SortAsc, opts.SortBy[0].Order)
}

func TestParseListOptions_ZeroBasedPagination(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		opts := ParseListOptions(c, "test-org")
		return c.JSON(opts)
	})

	// Test page 0 (first page)
	req := httptest.NewRequest(http.MethodGet, "/test?page=0&pageSize=10", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	var opts domain.ListOptions
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 0, opts.Pagination.Page, "Page 0 should be accepted")
	assert.Equal(t, 10, opts.Pagination.PageSize)

	// Test page 1 (second page)
	req = httptest.NewRequest(http.MethodGet, "/test?page=1&pageSize=10", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&opts)
	assert.NoError(t, err)
	assert.Equal(t, 1, opts.Pagination.Page, "Page 1 should be accepted")
	assert.Equal(t, 10, opts.Pagination.PageSize)
}

func TestEventTypeFilterFromQuery(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		filter := EventTypeFilterFromQuery(c)
		return c.JSON(map[string]string{"eventType": filter})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?eventType=%20order.created%20", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "order.created", body["eventType"])
}
