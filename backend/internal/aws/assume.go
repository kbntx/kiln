package aws

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Credentials holds temporary AWS credentials from an AssumeRole call.
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// AssumeRole calls STS AssumeRole using the server's own credentials
// (instance profile, task role, env vars) and returns temporary credentials.
func AssumeRole(ctx context.Context, roleARN, region, sessionName string) (Credentials, error) {
	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return Credentials{}, fmt.Errorf("load aws config: %w", err)
	}

	client := sts.NewFromConfig(cfg)
	out, err := client.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         &roleARN,
		RoleSessionName: &sessionName,
	})
	if err != nil {
		return Credentials{}, fmt.Errorf("assume role %s: %w", roleARN, err)
	}

	return Credentials{
		AccessKeyID:     *out.Credentials.AccessKeyId,
		SecretAccessKey: *out.Credentials.SecretAccessKey,
		SessionToken:    *out.Credentials.SessionToken,
	}, nil
}
