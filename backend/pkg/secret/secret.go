package secret

import (
	"context"
	"log"
	"path/filepath"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// Provider - handle secret manager
type Provider interface {
	GetSecret(ctx context.Context, name string) string
}

type provider struct {
	client    *secretmanager.Client
	projectID string
}

// NewProvider - constructor
func NewProvider(client *secretmanager.Client, projectID string) Provider {
	return &provider{
		client:    client,
		projectID: projectID,
	}
}

// GetSecret - get confidential information
func (p *provider) GetSecret(ctx context.Context, name string) string {
	secretPath := filepath.Join("projects", p.projectID, "secrets", name, "versions", "latest")
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	res, err := p.client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatalf("AccessSecretVersion error: %+v", err)
	}

	return string(res.Payload.Data)
}
