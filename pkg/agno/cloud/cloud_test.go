package cloud

import (
    "context"
    "testing"
)

func TestNoopDeployer(t *testing.T) {
    d := NoopDeployer{}
    id, err := d.Deploy(context.Background(), "artifact.bin", nil)
    if err != nil { t.Fatalf("Deploy error: %v", err) }
    if id != "local://artifact.bin" { t.Fatalf("unexpected id: %s", id) }
}

