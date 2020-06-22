package secrets

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
)

type SecretGetter interface {
	SecretToken(ctx context.Context, key types.NamespacedName) (string, error)
}
