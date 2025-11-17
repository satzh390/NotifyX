package middlewares

import "github.com/gofiber/fiber/v2"

const orgContextKey = "orgId"

// OrgIDFromCtx extracts the organization ID from Fiber context
func OrgIDFromCtx(fiberCtx *fiber.Ctx) string {
	if value := fiberCtx.Locals(orgContextKey); value != nil {
		if orgID, ok := value.(string); ok {
			return orgID
		}
	}
	return ""
}

