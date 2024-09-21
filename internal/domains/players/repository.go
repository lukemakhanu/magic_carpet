package players

import "context"

// PlayersRepository contains methods that implements players struct
type PlayersRepository interface {
	Save(ctx context.Context, t Players) (int, error)
	PlayerExists(ctx context.Context, profileTag string) ([]Players, error)
	UpdatePlayer(ctx context.Context, profileTag, status string) (int64, error)
}
