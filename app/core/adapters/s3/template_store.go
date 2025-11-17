package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

var (
	ErrNotFound = storage.ErrNotFound
)

type Options struct {
	Bucket      string
	Region      string
	Endpoint    string // For LocalStack
	KeyPrefix   string // Optional prefix for all keys
	AccessKeyID string // Optional, for LocalStack
	SecretKey   string // Optional, for LocalStack
}

type TemplateRepository struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewTemplateRepository creates a new S3-backed template repository.
func NewTemplateRepository(ctx context.Context, opts Options) (*TemplateRepository, error) {
	if opts.Bucket == "" {
		return nil, errors.New("s3: bucket is required")
	}
	if opts.Region == "" {
		opts.Region = "us-east-1"
	}

	cfg, err := loadAWSConfig(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("s3: config: %w", err)
	}

	clientOptions := []func(*s3.Options){}
	if opts.Endpoint != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(opts.Endpoint)
			o.UsePathStyle = true // Required for LocalStack
		})
	}

	client := s3.NewFromConfig(cfg, clientOptions...)

	prefix := strings.Trim(opts.KeyPrefix, "/")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return &TemplateRepository{
		client: client,
		bucket: opts.Bucket,
		prefix: prefix,
	}, nil
}

func loadAWSConfig(ctx context.Context, opts Options) (aws.Config, error) {
	cfgOptions := []func(*config.LoadOptions) error{
		config.WithRegion(opts.Region),
	}

	// Override credentials if provided (for LocalStack)
	if opts.AccessKeyID != "" && opts.SecretKey != "" {
		cfgOptions = append(cfgOptions, config.WithCredentialsProvider(
			aws.NewStaticCredentialsProvider(opts.AccessKeyID, opts.SecretKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, cfgOptions...)
	if err != nil {
		return aws.Config{}, err
	}

	return cfg, nil
}

// buildKey constructs the S3 key for a template.
// Format: {prefix}{orgId}/templates/{templateId}_{language}_{channel}.json
// Example: acme/templates/order_created_default_email.json
func (repo *TemplateRepository) buildKey(orgID, templateID, language, channel string) string {
	if language == "" {
		language = "default"
	}
	key := fmt.Sprintf("%s%s/templates/%s_%s_%s.json", repo.prefix, orgID, templateID, language, channel)
	return strings.TrimPrefix(key, "/")
}

func (repo *TemplateRepository) Put(ctx context.Context, template domain.Template) error {
	// Extract language from template metadata or use "default"
	language := template.Metadata["language"]
	if language == "" {
		language = "default"
	}

	key := repo.buildKey(template.OrgID, template.ID, language, string(template.Channel))

	// Convert template to JSON
	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("s3: marshal: %w", err)
	}

	_, err = repo.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(repo.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(templateJSON),
		ContentType: aws.String("application/json"),
	})

	return err
}

func (repo *TemplateRepository) Get(ctx context.Context, orgID, templateID string) (domain.Template, error) {
	// Try to get the default language first, then fallback to any language
	// For now, we'll try "default" language and common channels
	channels := []string{"email", "sms", "push", "webhook"}
	language := "default"

	for _, channel := range channels {
		key := repo.buildKey(orgID, templateID, language, channel)
		template, err := repo.getByKey(ctx, key)
		if err == nil {
			return template, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return domain.Template{}, err
		}
	}

	return domain.Template{}, ErrNotFound
}

// GetByLanguage retrieves a template with a specific language.
func (repo *TemplateRepository) GetByLanguage(ctx context.Context, orgID, templateID, language, channel string) (domain.Template, error) {
	if channel == "" {
		channel = "email" // default channel
	}
	key := repo.buildKey(orgID, templateID, language, channel)
	return repo.getByKey(ctx, key)
}

func (repo *TemplateRepository) getByKey(ctx context.Context, key string) (domain.Template, error) {
	result, err := repo.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(repo.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *s3.NoSuchKey
		if errors.As(err, &nsk) {
			return domain.Template{}, ErrNotFound
		}
		return domain.Template{}, fmt.Errorf("s3: get object: %w", err)
	}
	defer result.Body.Close()

	var template domain.Template
	if err := json.NewDecoder(result.Body).Decode(&template); err != nil {
		return domain.Template{}, fmt.Errorf("s3: decode: %w", err)
	}

	// Ensure timestamps are set if missing
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}
	if template.UpdatedAt.IsZero() {
		template.UpdatedAt = time.Now()
	}

	return template, nil
}

func (repo *TemplateRepository) Delete(ctx context.Context, orgID, templateID string) error {
	// Delete all variants of the template (all languages and channels)
	prefix := fmt.Sprintf("%s%s/templates/%s_", repo.prefix, orgID, templateID)
	
	// List all objects with this prefix
	paginator := s3.NewListObjectsV2Paginator(repo.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(repo.bucket),
		Prefix: aws.String(prefix),
	})

	var keysToDelete []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("s3: list objects: %w", err)
		}

		for _, obj := range page.Contents {
			keysToDelete = append(keysToDelete, *obj.Key)
		}
	}

	if len(keysToDelete) == 0 {
		return ErrNotFound
	}

	// Delete all objects
	for _, key := range keysToDelete {
		_, err := repo.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(repo.bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("s3: delete object %s: %w", key, err)
		}
	}

	return nil
}

