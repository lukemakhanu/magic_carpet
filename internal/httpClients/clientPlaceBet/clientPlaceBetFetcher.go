package clientPlaceBet

import (
	"context"

	"github.com/lukemakhanu/veimu_apps/internal/domain/clientBets"
)

// ClientInformationFetcher : returns client info
type ClientPlaceBetFetcher interface {
	SubmitBet(ctx context.Context, p clientBets.SubmitBetToClient) (*SubmitBetResponse, error)
}
