package clientInformation

import (
	"context"
)

// ClientInformationFetcher : returns client info
type ClientInformationFetcher interface {
	GetClientInfo(ctx context.Context) (*ClientInfoApi, error)
}
