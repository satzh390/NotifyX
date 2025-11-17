package utils

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
)

func TestMergePatch(t *testing.T) {
	tests := []struct {
		name     string
		target   interface{}
		patch    string
		expected interface{}
		wantErr  bool
	}{
		{
			name: "simple field update",
			target: &domain.Group{
				ID:          "group-1",
				OrgID:       "org-1",
				Name:        "Old Name",
				Description: "Old Description",
			},
			patch: `{"name": "New Name"}`,
			expected: &domain.Group{
				ID:          "group-1",
				OrgID:       "org-1",
				Name:        "New Name",
				Description: "Old Description",
			},
			wantErr: false,
		},
		{
			name: "multiple field update",
			target: &domain.Group{
				ID:          "group-1",
				OrgID:       "org-1",
				Name:        "Old Name",
				Description: "Old Description",
			},
			patch: `{"name": "New Name", "description": "New Description"}`,
			expected: &domain.Group{
				ID:          "group-1",
				OrgID:       "org-1",
				Name:        "New Name",
				Description: "New Description",
			},
			wantErr: false,
		},
		{
			name: "nested object merge",
			target: &domain.Subscriber{
				ID:    "sub-1",
				OrgID: "org-1",
				Preferences: domain.SubscriberPrefs{
					Language: "en",
					AllowedDays: []string{"monday", "tuesday"},
					NotificationWindow: domain.TimeWindow{
						Start: "09:00",
						End:   "17:00",
					},
				},
			},
			patch: `{"preferences": {"language": "fr", "notificationWindow": {"start": "10:00"}}}`,
			expected: &domain.Subscriber{
				ID:    "sub-1",
				OrgID: "org-1",
				Preferences: domain.SubscriberPrefs{
					Language:    "fr",
					AllowedDays: []string{"monday", "tuesday"},
					NotificationWindow: domain.TimeWindow{
						Start: "10:00",
						End:   "17:00", // Should be preserved
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty patch",
			target: &domain.Group{
				ID:   "group-1",
				Name: "Test",
			},
			patch:    ``,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid JSON",
			target: &domain.Group{
				ID:   "group-1",
				Name: "Test",
			},
			patch:    `{invalid json}`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergePatch(tt.target, []byte(tt.patch))
			if (err != nil) != tt.wantErr {
				t.Errorf("MergePatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Compare JSON representations
			gotJSON, _ := json.Marshal(tt.target)
			expectedJSON, _ := json.Marshal(tt.expected)
			if string(gotJSON) != string(expectedJSON) {
				t.Errorf("MergePatch() = %v, want %v", string(gotJSON), string(expectedJSON))
			}
		})
	}
}

func TestMergePatchWithTime(t *testing.T) {
	now := time.Now()
	target := &domain.Subscriber{
		ID:        "sub-1",
		OrgID:     "org-1",
		CreatedAt: now,
	}

	patch := `{"email": "test@example.com"}`
	err := MergePatch(target, []byte(patch))
	if err != nil {
		t.Fatalf("MergePatch() error = %v", err)
	}

	if target.Email != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", target.Email)
	}
	if !target.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt was modified, want %v", now)
	}
}

