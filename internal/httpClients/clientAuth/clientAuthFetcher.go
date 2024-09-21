package clientAuth

import (
	"context"
)

// ClientAuthFetcher : returns client auth
type ClientAuthFetcher interface {
	GetClientAuth(ctx context.Context) (*ClientAuth, error)
}
