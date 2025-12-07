package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// SNSProvider implements SMS sending via AWS SNS
type SNSProvider struct {
	client *sns.Client
	region string
}

type SNSConfig struct {
	Region    string
	AccessKey string // Optional, uses default credentials if empty
	SecretKey string // Optional, uses default credentials if empty
	Endpoint  string // Optional, for LocalStack or custom endpoints
}

func NewSNSProvider(cfg SNSConfig) (*SNSProvider, error) {
	var awsCfg aws.Config
	var err error

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		// Use explicit credentials
		awsCfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		)
	} else {
		// Use default credentials (IAM role, env vars, etc.)
		awsCfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.Region),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("sms: sns config: %w", err)
	}

	// Create SNS client with optional custom endpoint (for LocalStack)
	opts := []func(*sns.Options){}
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *sns.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	return &SNSProvider{
		client: sns.NewFromConfig(awsCfg, opts...),
		region: cfg.Region,
	}, nil
}

func (p *SNSProvider) Send(ctx context.Context, phone, message string, metadata map[string]string) error {
	input := &sns.PublishInput{
		PhoneNumber: aws.String(phone),
		Message:     aws.String(message),
	}

	// Add metadata as message attributes if provided
	if len(metadata) > 0 {
		attrs := make(map[string]types.MessageAttributeValue)
		for k, v := range metadata {
			attrs[k] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(v),
			}
		}
		input.MessageAttributes = attrs
	}

	_, err := p.client.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("sms: sns publish: %w", err)
	}

	return nil
}

func (p *SNSProvider) Name() string {
	return "aws-sns"
}
