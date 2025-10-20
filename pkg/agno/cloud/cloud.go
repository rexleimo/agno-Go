package cloud

import (
    "context"
)

// Deployer defines a minimal deployment interface
type Deployer interface {
    // Deploy publishes an artifact and returns a deployment ID or URL
    Deploy(ctx context.Context, artifact string, config map[string]string) (string, error)
}

// NoopDeployer is a placeholder deployer for local/testing usage
type NoopDeployer struct{}

func (NoopDeployer) Deploy(_ context.Context, artifact string, _ map[string]string) (string, error) {
    // Return a pseudo-URL to indicate deployment location
    return "local://" + artifact, nil
}

