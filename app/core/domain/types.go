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
	ID          string            `json:"subscriberId" bson:"subscriberId"`
	OrgID       string            `json:"orgId" bson:"orgId"`
	Email       string            `json:"email,omitempty" bson:"email,omitempty"`
	Phone       string            `json:"phone,omitempty" bson:"phone,omitempty"`
	PushToken   string            `json:"pushToken,omitempty" bson:"pushToken,omitempty"`
	WebhookURL  string            `json:"webhookUrl,omitempty" bson:"webhookUrl,omitempty"`
	Preferences SubscriberPrefs   `json:"preferences" bson:"preferences"`
	Groups      []string          `json:"groups" bson:"groups"`
	Metadata    map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt" bson:"createdAt"`
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
	ID          string            `json:"groupId" bson:"groupId"`
	OrgID       string            `json:"orgId" bson:"orgId"`
	Name        string            `json:"name" bson:"name"`
	Description string            `json:"description,omitempty" bson:"description,omitempty"`
	Subscribers []string          `json:"subscribers" bson:"subscribers"`
	Metadata    map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type Template struct {
	ID        string            `json:"id" bson:"id"`
	OrgID     string            `json:"orgId" bson:"orgId"`
	Name      string            `json:"name" bson:"name"`
	Channel   ChannelType       `json:"channel" bson:"channel"`
	Version   int               `json:"version" bson:"version"`
	Body      string            `json:"body" bson:"body"`
	Subject   string            `json:"subject,omitempty" bson:"subject,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt" bson:"updatedAt"`
}

type Rule struct {
	EventType         string                 `json:"eventType" bson:"eventType"`
	OrgID             string                 `json:"orgId" bson:"orgId"`
	Channels          []ChannelType          `json:"channels" bson:"channels"`
	DefaultRecipients Recipients             `json:"defaultRecipients" bson:"defaultRecipients"`
	TemplateRefs      map[ChannelType]string `json:"templateRefs" bson:"templateRefs"`
	CreatedAt         time.Time              `json:"createdAt" bson:"createdAt"`
}

type Event struct {
	ID         string            `json:"eventId" bson:"eventId"`
	OrgID      string            `json:"orgId" bson:"orgId"`
	Type       string            `json:"type" bson:"type"`
	Recipients Recipients        `json:"recipients" bson:"recipients"`
	Payload    map[string]any    `json:"payload" bson:"payload"`
	Meta       map[string]string `json:"meta,omitempty" bson:"meta,omitempty"`
}

type Recipients struct {
	SubscriberIDs []string `json:"subscriberIds" bson:"subscriberIds"`
	Groups        []string `json:"groups" bson:"groups"`
	Broadcast     bool     `json:"broadcast" bson:"broadcast"`
}

type DeliveryTask struct {
	TaskID      string            `json:"taskId" bson:"taskId"`
	EventID     string            `json:"eventId" bson:"eventId"`
	OrgID       string            `json:"orgId" bson:"orgId"`
	Subscriber  Subscriber        `json:"subscriber" bson:"subscriber"`
	Channel     ChannelType       `json:"channel" bson:"channel"`
	TemplateRef string            `json:"templateRef" bson:"templateRef"`
	Payload     map[string]any    `json:"payload" bson:"payload"`
	Metadata    map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt" bson:"createdAt"`
}

type DeliveryLog struct {
	TaskID    string            `json:"taskId" bson:"taskId"`
	OrgID     string            `json:"orgId" bson:"orgId"`
	EventID   string            `json:"eventId" bson:"eventId"`
	Channel   ChannelType       `json:"channel" bson:"channel"`
	Status    EventStatus       `json:"status" bson:"status"`
	Error     string            `json:"error,omitempty" bson:"error,omitempty"`
	Timestamp time.Time         `json:"timestamp" bson:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// Pagination types
type PaginationParams struct {
	Page     int `json:"page"`     // 1-based page number
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
type SortOption struct {
	Field string    `json:"field"` // Field name to sort by
	Order SortOrder `json:"order"` // Sort direction (asc/desc)
}

// ListOptions contains pagination, filtering, and sorting options
type ListOptions struct {
	Pagination PaginationParams
	OrgID      string       // Organization filter (required)
	GroupID    string       // Group filter (for subscribers only)
	SortBy     []SortOption // Sort options (multiple fields supported)
}

// ListResult contains paginated results
type ListResult[T any] struct {
	Items      []T              `json:"items"`
	Pagination PaginationResult `json:"pagination"`
}

