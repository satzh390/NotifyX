package domain

import "time"

type (
	ChannelType string
	EventStatus string
)

const (
	ChannelEmail   ChannelType = "email"
	ChannelSMS     ChannelType = "sms"
	ChannelPush    ChannelType = "push"
	ChannelWebhook ChannelType = "webhook"

	EventStatusPending   EventStatus = "pending"
	EventStatusFanout    EventStatus = "fanout"
	EventStatusDelivered EventStatus = "delivered"
	EventStatusFailed    EventStatus = "failed"
)

type Subscriber struct {
	ID                   string            `json:"subscriberId" bson:"subscriberId" immutable:"true"`
	CustomerID           string            `json:"customerId" bson:"customerId" immutable:"true"`
	Email                string            `json:"email,omitempty" bson:"email,omitempty"`
	Phone                string            `json:"phone,omitempty" bson:"phone,omitempty"`
	PushToken            string            `json:"pushToken,omitempty" bson:"pushToken,omitempty"`   // Deprecated: Use PushTokens instead
	PushTokens           map[string]string `json:"pushTokens,omitempty" bson:"pushTokens,omitempty"` // Map of appId -> push token
	WebhookURL           string            `json:"webhookUrl,omitempty" bson:"webhookUrl,omitempty"`
	Preferences          SubscriberPrefs   `json:"preferences" bson:"preferences"`
	Groups               []string          `json:"groups" bson:"groups"`
	SubscribedEventTypes []string          `json:"subscribedEventTypes,omitempty" bson:"subscribedEventTypes,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt            time.Time         `json:"createdAt" bson:"createdAt" immutable:"true"`
}

type SubscriberPrefs struct {
	Channels           map[ChannelType]bool `json:"channels" bson:"channels"`
	Language           string               `json:"language" bson:"language"`
	AllowedDays        []string             `json:"allowedDays" bson:"allowedDays"`
	NotificationWindow TimeWindow           `json:"notificationWindow" bson:"notificationWindow"`
}

type TimeWindow struct {
	Start string `json:"start" bson:"start"`
	End   string `json:"end" bson:"end"`
}

type Group struct {
	ID                   string            `json:"groupId" bson:"groupId" immutable:"true"`
	CustomerID           string            `json:"customerId" bson:"customerId" immutable:"true"`
	Name                 string            `json:"name" bson:"name"`
	Description          string            `json:"description,omitempty" bson:"description,omitempty"`
	Subscribers          []string          `json:"subscribers" bson:"subscribers"`
	SubscribedEventTypes []string          `json:"subscribedEventTypes,omitempty" bson:"subscribedEventTypes,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// TemplateContent represents channel-specific template content
type TemplateContent struct {
	// Email content
	Subject string `json:"subject,omitempty" bson:"subject,omitempty"`
	Body    string `json:"body" bson:"body"`

	// SMS content (only body)
	// Body is reused for SMS

	// Push notification content
	Title string `json:"title,omitempty" bson:"title,omitempty"`
	// Body is reused for push body

	// Webhook content (arbitrary JSON)
	Payload map[string]any `json:"payload,omitempty" bson:"payload,omitempty"`
}

// Template represents a notification template with channel-specific content and translations
type Template struct {
	ID           string                     `json:"id" bson:"id" immutable:"true"`
	CustomerID   string                     `json:"customerId" bson:"customerId" immutable:"true"`
	Name         string                     `json:"name" bson:"name"`
	Channel      ChannelType                `json:"channel" bson:"channel" immutable:"true"`
	Version      int                        `json:"version" bson:"version"`
	Content      TemplateContent            `json:"content" bson:"content"`
	Translations map[string]TemplateContent `json:"translations,omitempty" bson:"translations,omitempty"` // key: language code (e.g., "en", "es", "fr")
	Metadata     map[string]string          `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt    time.Time                  `json:"createdAt" bson:"createdAt" immutable:"true"`
	UpdatedAt    time.Time                  `json:"updatedAt" bson:"updatedAt"`
}

// CustomFilterConfig represents configuration for a custom filter.
// The Type field must match the name of a registered custom filter (typically the struct type name).
// Custom filters are applied in addition to the default preferences filter and are NOT applied to direct emails and phone numbers.
type CustomFilterConfig struct {
	Type   string                 `json:"type" bson:"type"`     // Filter type/name identifier (must match registered filter name)
	Config map[string]interface{} `json:"config" bson:"config"` // Filter-specific configuration passed to the filter
}

type Rule struct {
	EventType    string                 `json:"eventType" bson:"eventType" immutable:"true"`
	CustomerID   string                 `json:"customerId" bson:"customerId" immutable:"true"`
	Channels     []ChannelType          `json:"channels" bson:"channels"`
	TemplateRefs map[ChannelType]string `json:"templateRefs" bson:"templateRefs"`
	CustomFilter *CustomFilterConfig    `json:"customFilter,omitempty" bson:"customFilter,omitempty"` // Optional custom filter configuration
	Metadata     map[string]string      `json:"metadata,omitempty" bson:"metadata,omitempty"`         // Rule metadata (e.g., appId for push notifications)
	CreatedAt    time.Time              `json:"createdAt" bson:"createdAt" immutable:"true"`
	UpdatedAt    time.Time              `json:"updatedAt" bson:"updatedAt"`
}

type Event struct {
	ID         string            `json:"eventId" bson:"eventId"`
	CustomerID string            `json:"customerId" bson:"customerId"`
	Type       string            `json:"type" bson:"type"`
	Recipients Recipients        `json:"recipients" bson:"recipients"`
	Payload    map[string]any    `json:"payload" bson:"payload"`
	Meta       map[string]string `json:"meta,omitempty" bson:"meta,omitempty"`
}

type Recipients struct {
	SubscriberIDs []string `json:"subscriberIds" bson:"subscriberIds" validate:"max=5000"`
	Groups        []string `json:"groups" bson:"groups" validate:"max=100"`
	Broadcast     bool     `json:"broadcast" bson:"broadcast"`
	DirectEmails  []string `json:"directEmails,omitempty" bson:"directEmails,omitempty" validate:"max=5000"`
	DirectPhones  []string `json:"directPhones,omitempty" bson:"directPhones,omitempty" validate:"max=5000"`
}

type DeliveryTask struct {
	TaskID         string            `json:"taskId" bson:"taskId"`
	IdempotencyKey string            `json:"idempotencyKey" bson:"idempotencyKey"`
	EventID        string            `json:"eventId" bson:"eventId"`
	CustomerID     string            `json:"customerId" bson:"customerId"`
	Subscriber     Subscriber        `json:"subscriber" bson:"subscriber"`
	Channel        ChannelType       `json:"channel" bson:"channel"`
	TemplateRef    string            `json:"templateRef" bson:"templateRef"`
	Payload        map[string]any    `json:"payload" bson:"payload"`
	Metadata       map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"createdAt" bson:"createdAt"`
}

type DeliveryLog struct {
	TaskID     string            `json:"taskId" bson:"taskId"`
	CustomerID string            `json:"customerId" bson:"customerId"`
	EventID    string            `json:"eventId" bson:"eventId"`
	Channel    ChannelType       `json:"channel" bson:"channel"`
	Status     EventStatus       `json:"status" bson:"status"`
	Error      string            `json:"error,omitempty" bson:"error,omitempty"`
	Timestamp  time.Time         `json:"timestamp" bson:"timestamp"`
	Metadata   map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type PaginationParams struct {
	Page     int `json:"page"`     // 0-based page number
	PageSize int `json:"pageSize"` // Number of items per page
}

type PaginationResult struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalCount int64 `json:"totalCount"`
	TotalPages int   `json:"totalPages"`
}

// SortOrder represents sort direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// SortOption represents a single sort field and direction
type SortParams struct {
	Field string    `json:"field"` // Field name to sort by
	Order SortOrder `json:"order"` // Sort direction (asc/desc)
}

// PageAndSortOption contains pagination and sorting options
type ListOptions struct {
	Pagination PaginationParams
	SortBy     []SortParams // Sort options (multiple fields supported)
	Filter     map[string]string
}

// ListResult contains paginated results
type ListResult[T any] struct {
	Items      []T              `json:"items"`
	Pagination PaginationResult `json:"pagination"`
}

// OrganizationType represents the type of organization
type OrganizationType string

const (
	OrganizationTypeCompany      OrganizationType = "company"
	OrganizationTypeSaaSProvider OrganizationType = "saasProvider"
)

// Organization represents an organization entity
type Organization struct {
	ID        string           `json:"id" bson:"id" immutable:"true"`
	Name      string           `json:"name" bson:"name"`
	Type      OrganizationType `json:"type" bson:"type"`
	CreatedAt time.Time        `json:"createdAt" bson:"createdAt" immutable:"true"`
	UpdatedAt time.Time        `json:"updatedAt" bson:"updatedAt"`
}

// Customer represents a customer entity
type Customer struct {
	ID        string            `json:"id" bson:"id" immutable:"true"`
	OrgID     string            `json:"orgId" bson:"orgId" immutable:"true"`
	Name      string            `json:"name" bson:"name"`
	Logo      string            `json:"logo,omitempty" bson:"logo,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt time.Time         `json:"createdAt" bson:"createdAt" immutable:"true"`
	UpdatedAt time.Time         `json:"updatedAt" bson:"updatedAt"`
}

// PushProviderType represents the type of push notification provider
type PushProviderType string

const (
	PushProviderAPNS     PushProviderType = "apns"
	PushProviderFirebase PushProviderType = "firebase"
	PushProviderMock     PushProviderType = "mock"
)

// AppConfig represents push notification configuration for an app, bound to an Organization
type AppConfig struct {
	ID        string            `json:"id" bson:"id" immutable:"true"`
	OrgID     string            `json:"orgId" bson:"orgId" immutable:"true"`
	Name      string            `json:"name" bson:"name"`                             // App name/identifier
	Provider  PushProviderType  `json:"provider" bson:"provider"`                     // "apns", "firebase", or "mock"
	APNS      *APNSConfig       `json:"apns,omitempty" bson:"apns,omitempty"`         // APNS configuration (if provider is "apns")
	Firebase  *FirebaseConfig   `json:"firebase,omitempty" bson:"firebase,omitempty"` // Firebase configuration (if provider is "firebase")
	Metadata  map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt time.Time         `json:"createdAt" bson:"createdAt" immutable:"true"`
	UpdatedAt time.Time         `json:"updatedAt" bson:"updatedAt"`
}

// APNSConfig represents APNS provider configuration
type APNSConfig struct {
	KeyID      string `json:"keyId" bson:"keyId"`
	TeamID     string `json:"teamId" bson:"teamId"`
	BundleID   string `json:"bundleId" bson:"bundleId"`
	KeyPath    string `json:"keyPath" bson:"keyPath"` // Path to APNS key file (.p8)
	Production bool   `json:"production" bson:"production"`
}

// FirebaseConfig represents Firebase provider configuration
type FirebaseConfig struct {
	ProjectID  string `json:"projectId" bson:"projectId"`
	Credential string `json:"credential" bson:"credential"` // Path to Firebase service account JSON file
}
