package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"go.mongodb.org/mongo-driver/bson"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TemplateRepository struct {
	collection *mongoDriver.Collection
}

func (repo *TemplateRepository) Put(ctx context.Context, template domain.Template) error {
	// Template ID is unique per org, channel is a property of the template
	filter := bson.M{
		"orgId": template.OrgID,
		"id":    template.ID,
	}
	updateMap, err := BuildUpdateMap(template)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"orgId":     template.OrgID,
			"id":        template.ID,
			"channel":   template.Channel,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *TemplateRepository) Get(ctx context.Context, orgID, templateID string) (domain.Template, error) {
	filter := bson.M{
		"orgId": orgID,
		"id":    templateID,
	}
	var template domain.Template
	err := repo.collection.FindOne(ctx, filter).Decode(&template)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.Template{}, storage.ErrNotFound
	}

	if err != nil {
		return domain.Template{}, err
	}

	return template, nil
}

func (repo *TemplateRepository) GetByLanguage(ctx context.Context, orgID, templateID, language string) (domain.Template, error) {
	template, err := repo.Get(ctx, orgID, templateID)
	if err != nil {
		return domain.Template{}, err
	}

	// If language is specified and translation exists, merge it into content
	if language != "" && template.Translations != nil {
		if translatedContent, ok := template.Translations[language]; ok {
			// Merge translated content into the main content
			// Translated content takes precedence
			if translatedContent.Body != "" {
				template.Content.Body = translatedContent.Body
			}
			if translatedContent.Subject != "" {
				template.Content.Subject = translatedContent.Subject
			}
			if translatedContent.Title != "" {
				template.Content.Title = translatedContent.Title
			}
			if translatedContent.Payload != nil {
				template.Content.Payload = translatedContent.Payload
			}
		}
	}

	return template, nil
}

func (repo *TemplateRepository) Delete(ctx context.Context, orgID, templateID string) error {
	// Delete all variants (all channels and languages)
	filter := bson.M{
		"orgId": orgID,
		"id":    templateID,
	}
	result, err := repo.collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

