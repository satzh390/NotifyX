package handlers

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

// Mock stores for testing
type mockSubscriberStore struct {
	subscribers map[string]domain.Subscriber
	lists       []domain.ListResult[domain.Subscriber]
	listErr     error
}

func (m *mockSubscriberStore) Put(ctx context.Context, sub domain.Subscriber) error {
	if m.subscribers == nil {
		m.subscribers = make(map[string]domain.Subscriber)
	}
	m.subscribers[sub.ID] = sub
	return nil
}

func (m *mockSubscriberStore) Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, error) {
	sub, ok := m.subscribers[subscriberID]
	if !ok || sub.OrgID != orgID {
		return domain.Subscriber{}, storage.ErrNotFound
	}
	return sub, nil
}

func (m *mockSubscriberStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error) {
	if m.listErr != nil {
		return domain.ListResult[domain.Subscriber]{}, m.listErr
	}
	if len(m.lists) > 0 {
		return m.lists[0], nil
	}
	return domain.ListResult[domain.Subscriber]{
		Items: []domain.Subscriber{},
		Pagination: domain.PaginationResult{
			Page:       opts.Pagination.Page,
			PageSize:   opts.Pagination.PageSize,
			TotalCount: 0,
			TotalPages: 0,
		},
	}, nil
}

func (m *mockSubscriberStore) Delete(ctx context.Context, orgID, subscriberID string) error {
	sub, ok := m.subscribers[subscriberID]
	if !ok || sub.OrgID != orgID {
		return storage.ErrNotFound
	}
	delete(m.subscribers, subscriberID)
	return nil
}

type mockGroupStore struct {
	groups map[string]domain.Group
}

func (m *mockGroupStore) Put(ctx context.Context, group domain.Group) error {
	if m.groups == nil {
		m.groups = make(map[string]domain.Group)
	}
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupStore) Get(ctx context.Context, orgID, groupID string) (domain.Group, error) {
	group, ok := m.groups[groupID]
	if !ok || group.OrgID != orgID {
		return domain.Group{}, storage.ErrNotFound
	}
	return group, nil
}

func (m *mockGroupStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error) {
	var items []domain.Group
	for _, g := range m.groups {
		if g.OrgID == opts.OrgID {
			items = append(items, g)
		}
	}
	return domain.ListResult[domain.Group]{
		Items: items,
		Pagination: domain.PaginationResult{
			Page:       opts.Pagination.Page,
			PageSize:   opts.Pagination.PageSize,
			TotalCount: int64(len(items)),
			TotalPages: 1,
		},
	}, nil
}

func (m *mockGroupStore) Delete(ctx context.Context, orgID, groupID string) error {
	group, ok := m.groups[groupID]
	if !ok || group.OrgID != orgID {
		return storage.ErrNotFound
	}
	delete(m.groups, groupID)
	return nil
}

type mockRuleStore struct {
	rules map[string]domain.Rule
}

func (m *mockRuleStore) Put(ctx context.Context, rule domain.Rule) error {
	if m.rules == nil {
		m.rules = make(map[string]domain.Rule)
	}
	m.rules[rule.EventType] = rule
	return nil
}

func (m *mockRuleStore) Get(ctx context.Context, orgID, eventType string) (domain.Rule, error) {
	rule, ok := m.rules[eventType]
	if !ok || rule.OrgID != orgID {
		return domain.Rule{}, storage.ErrNotFound
	}
	return rule, nil
}

func (m *mockRuleStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error) {
	var items []domain.Rule
	for _, r := range m.rules {
		if r.OrgID == opts.OrgID {
			items = append(items, r)
		}
	}
	return domain.ListResult[domain.Rule]{
		Items: items,
		Pagination: domain.PaginationResult{
			Page:       opts.Pagination.Page,
			PageSize:   opts.Pagination.PageSize,
			TotalCount: int64(len(items)),
			TotalPages: 1,
		},
	}, nil
}

func (m *mockRuleStore) Delete(ctx context.Context, orgID, eventType string) error {
	rule, ok := m.rules[eventType]
	if !ok || rule.OrgID != orgID {
		return storage.ErrNotFound
	}
	delete(m.rules, eventType)
	return nil
}

type mockTemplateStore struct {
	templates map[string]domain.Template
}

func (m *mockTemplateStore) Put(ctx context.Context, template domain.Template) error {
	if m.templates == nil {
		m.templates = make(map[string]domain.Template)
	}
	m.templates[template.ID] = template
	return nil
}

func (m *mockTemplateStore) Get(ctx context.Context, orgID, templateID string) (domain.Template, error) {
	template, ok := m.templates[templateID]
	if !ok || template.OrgID != orgID {
		return domain.Template{}, storage.ErrNotFound
	}
	return template, nil
}

func (m *mockTemplateStore) Delete(ctx context.Context, orgID, templateID string) error {
	template, ok := m.templates[templateID]
	if !ok || template.OrgID != orgID {
		return storage.ErrNotFound
	}
	delete(m.templates, templateID)
	return nil
}

