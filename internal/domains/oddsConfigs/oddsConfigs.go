package oddsConfigs

import (
	"context"

	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
)

// OddsConfigsRepository : used to manipulate odds.
type OddsConfigsRepository interface {
	FormulateOdds(ctx context.Context) ([]oddsFiles.FinalMarkets, oddsFiles.FinalScores, []oddsFiles.FinalLiveScores, error)
}
