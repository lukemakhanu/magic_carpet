package clientProcessBet

import (
	"context"

	"github.com/lukemakhanu/veimu_apps/internal/domain/clientBets"
)

// ClientProcessBetFetcher : returns client info
type ClientProcessBetFetcher interface {
	SubmitWonBet(ctx context.Context, p clientBets.SubmitBetToClient) (*SubmitBetResponse, error)
}
