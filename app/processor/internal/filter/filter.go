package filter

import (
	"strings"
	"time"

	"github.com/notifyx/core/domain"
)

type FilteredSubscriber struct {
	Subscriber domain.Subscriber
	Channels   []domain.ChannelType
}

// SubscriberFilter defines the interface for filtering subscribers.
// The message parameter contains the event payload for custom filter evaluation.
type SubscriberFilter interface {
	Apply(subscribers []domain.Subscriber, rule domain.Rule, message map[string]interface{}) []FilteredSubscriber
}

type PreferencesFilter struct {
	now func() time.Time
}

func NewPreferencesFilter() *PreferencesFilter {
	return &PreferencesFilter{now: time.Now}
}

func (preferencesFilter *PreferencesFilter) Apply(subscribers []domain.Subscriber, rule domain.Rule, message map[string]interface{}) []FilteredSubscriber {
	now := preferencesFilter.now().UTC()
	result := make([]FilteredSubscriber, 0, len(subscribers))
	ruleChannels := map[domain.ChannelType]struct{}{}
	for _, ch := range rule.Channels {
		ruleChannels[ch] = struct{}{}
	}

	for _, sub := range subscribers {
		// Direct recipients (empty ID) bypass DND and time window checks
		isDirect := sub.ID == ""
		if !isDirect {
			if !allowedDay(sub, now) || !withinWindow(sub, now) {
				continue
			}
		}

		channels := intersectChannels(sub, ruleChannels, isDirect)
		if len(channels) == 0 {
			continue
		}
		result = append(result, FilteredSubscriber{
			Subscriber: sub,
			Channels:   channels,
		})
	}
	return result
}

func allowedDay(sub domain.Subscriber, now time.Time) bool {
	if len(sub.Preferences.AllowedDays) == 0 {
		return true
	}
	today := strings.ToLower(now.Weekday().String()[:3])
	for _, day := range sub.Preferences.AllowedDays {
		day = strings.ToLower(day)
		if len(day) >= 3 && day[:3] == today {
			return true
		}
	}
	return false
}

func withinWindow(sub domain.Subscriber, now time.Time) bool {
	window := sub.Preferences.NotificationWindow
	if window.Start == "" && window.End == "" {
		return true
	}

	start, err := time.Parse("15:04", window.Start)
	if err != nil {
		return true
	}
	end, err := time.Parse("15:04", window.End)
	if err != nil {
		return true
	}

	currentMinutes := now.Hour()*60 + now.Minute()
	startMinutes := start.Hour()*60 + start.Minute()
	endMinutes := end.Hour()*60 + end.Minute()

	if endMinutes < startMinutes {
		// window wraps midnight
		return currentMinutes >= startMinutes || currentMinutes <= endMinutes
	}
	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

func intersectChannels(sub domain.Subscriber, ruleChannels map[domain.ChannelType]struct{}, isDirect bool) []domain.ChannelType {
	if len(ruleChannels) == 0 {
		return nil
	}

	var channels []domain.ChannelType
	for ch := range ruleChannels {
		// Direct recipients only support email and SMS, not push
		if isDirect && ch == domain.ChannelPush {
			continue
		}
		if allowed, ok := sub.Preferences.Channels[ch]; ok && allowed {
			channels = append(channels, ch)
		}
	}
	return channels
}
