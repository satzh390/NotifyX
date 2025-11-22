package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/handlers/group"
	"github.com/notifyx/api/internal/handlers/rule"
	"github.com/notifyx/api/internal/handlers/subscriber"
	"github.com/notifyx/api/internal/handlers/template"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/httpx"
)

const (
	notifyReadPermission  = "notify:read"
	notifyWritePermission = "notify:write"
)

func RegisterRoutes(app *fiber.App, stores storage.Stores, validator httpx.AuthValidator) {
	api := app.Group("/api/v1")

	// Subscribers CRUD
	subscriberHandler := subscriber.NewSubscriberHandler(stores.Subscribers)
	subscribers := api.Group("/subscribers", httpx.RequireAuth(validator, notifyReadPermission))
	subscribers.Post("",
		httpx.RequireAuth(validator, notifyWritePermission),
		subscriberHandler.Create)
	subscribers.Get("/:id", subscriberHandler.Get)
	subscribers.Put("/:id",
		httpx.RequireAuth(validator, notifyWritePermission),
		subscriberHandler.Update)
	subscribers.Delete("/:id", httpx.RequireAuth(validator, notifyWritePermission), subscriberHandler.Delete)
	subscribers.Get("", subscriberHandler.List)

	// Groups CRUD
	groupHandler := group.NewGroupHandler(stores.Groups)
	groups := api.Group("/groups", httpx.RequireAuth(validator, notifyReadPermission))
	groups.Post("",
		httpx.RequireAuth(validator, notifyWritePermission),
		groupHandler.Create)
	groups.Get("/:id", groupHandler.Get)
	groups.Put("/:id",
		httpx.RequireAuth(validator, notifyWritePermission),
		groupHandler.Update)
	groups.Delete("/:id", httpx.RequireAuth(validator, notifyWritePermission), groupHandler.Delete)
	groups.Get("", groupHandler.List)

	// Rules CRUD
	ruleHandler := rule.NewRuleHandler(stores.Rules)
	rules := api.Group("/rules", httpx.RequireAuth(validator, notifyReadPermission))
	rules.Post("",
		httpx.RequireAuth(validator, notifyWritePermission),
		ruleHandler.Create)
	rules.Get("/:eventType", ruleHandler.Get)
	rules.Put("/:eventType",
		httpx.RequireAuth(validator, notifyWritePermission),
		ruleHandler.Update)
	rules.Delete("/:eventType", httpx.RequireAuth(validator, notifyWritePermission), ruleHandler.Delete)
	rules.Get("", ruleHandler.List)

	// Templates CRUD
	templateHandler := template.NewTemplateHandler(stores.Templates)
	templates := api.Group("/templates", httpx.RequireAuth(validator, notifyReadPermission))
	templates.Post("",
		httpx.RequireAuth(validator, notifyWritePermission),
		templateHandler.Create)
	templates.Get("/:id", templateHandler.Get)
	templates.Put("/:id",
		httpx.RequireAuth(validator, notifyWritePermission),
		templateHandler.Update)
	templates.Delete("/:id", httpx.RequireAuth(validator, notifyWritePermission), templateHandler.Delete)
}
