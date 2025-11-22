package rule

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRuleTestApp() (*fiber.App, *MockRuleStore) {
	app := fiber.New()
	store := new(MockRuleStore)
	handler := NewRuleHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org")
		return c.Next()
	})

	rules := api.Group("/rules")
	rules.Post("", handler.Create)
	rules.Get("/:eventType", handler.Get)
	rules.Put("/:eventType", handler.Update)
	rules.Delete("/:eventType", handler.Delete)
	rules.Get("", handler.List)

	return app, store
}

func TestRuleHandler_Create(t *testing.T) {
	app, store := setupRuleTestApp()

	store.On("Put", mock.Anything, mock.MatchedBy(func(r domain.Rule) bool {
		return r.EventType == "order.created" && r.OrgID == "test-org"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, "test-org", "order.created").Return(domain.Rule{
		EventType: "order.created",
		OrgID:     "test-org",
		Channels:  []domain.ChannelType{domain.ChannelEmail, domain.ChannelSMS},
	}, nil).Once()

	body := map[string]interface{}{
		"eventType": "order.created",
		"channels":  []string{"email", "sms"},
		"defaultRecipients": map[string]interface{}{
			"subscriberIds": []string{"sub-1"},
		},
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rules", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var rule domain.Rule
	err = json.NewDecoder(resp.Body).Decode(&rule)
	assert.NoError(t, err)
	assert.Equal(t, "order.created", rule.EventType)
	assert.Equal(t, "test-org", rule.OrgID)

	store.AssertExpectations(t)
}

func TestRuleHandler_Get(t *testing.T) {
	app, store := setupRuleTestApp()

	eventType := "order.created"
	expectedRule := domain.Rule{
		EventType: eventType,
		OrgID:     "test-org",
		Channels:  []domain.ChannelType{domain.ChannelEmail},
	}

	store.On("Get", mock.Anything, "test-org", eventType).Return(expectedRule, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules/"+eventType, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var rule domain.Rule
	err = json.NewDecoder(resp.Body).Decode(&rule)
	assert.NoError(t, err)
	assert.Equal(t, eventType, rule.EventType)

	store.AssertExpectations(t)
}

func TestRuleHandler_Get_NotFound(t *testing.T) {
	app, store := setupRuleTestApp()

	store.On("Get", mock.Anything, "test-org", "non-existent").Return(domain.Rule{}, storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestRuleHandler_Update(t *testing.T) {
	app, store := setupRuleTestApp()

	eventType := "order.created"
	existingRule := domain.Rule{
		EventType: eventType,
		OrgID:     "test-org",
		Channels:  []domain.ChannelType{domain.ChannelEmail},
	}

	store.On("Get", mock.Anything, "test-org", eventType).Return(existingRule, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(r domain.Rule) bool {
		return r.EventType == eventType && len(r.Channels) == 2
	})).Return(nil).Once()
	store.On("Get", mock.Anything, "test-org", eventType).Return(domain.Rule{
		EventType: eventType,
		OrgID:     "test-org",
		Channels:  []domain.ChannelType{domain.ChannelEmail, domain.ChannelSMS},
	}, nil).Once()

	patch := map[string]interface{}{
		"channels": []string{"email", "sms"},
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/rules/"+eventType, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestRuleHandler_Delete(t *testing.T) {
	app, store := setupRuleTestApp()

	eventType := "order.created"
	existingRule := domain.Rule{
		EventType: eventType,
		OrgID:     "test-org",
	}

	store.On("Get", mock.Anything, "test-org", eventType).Return(existingRule, nil).Once()
	store.On("Delete", mock.Anything, "test-org", eventType).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/rules/"+eventType, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestRuleHandler_List(t *testing.T) {
	app, store := setupRuleTestApp()

	expectedResult := domain.ListResult[domain.Rule]{
		Items: []domain.Rule{
			{EventType: "order.created", OrgID: "test-org"},
			{EventType: "order.updated", OrgID: "test-org"},
		},
		Pagination: domain.PaginationResult{
			Page:       0,
			PageSize:   20,
			TotalCount: 2,
			TotalPages: 1,
		},
	}

	store.On("List", mock.Anything, mock.MatchedBy(func(opts domain.ListOptions) bool {
		return opts.Pagination.Page == 0 && opts.Pagination.PageSize == 20
	})).Return(expectedResult, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules?page=0&pageSize=20", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result domain.ListResult[domain.Rule]
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Pagination.TotalCount, int64(2))

	store.AssertExpectations(t)
}
