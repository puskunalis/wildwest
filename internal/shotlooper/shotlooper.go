package shotlooper

import (
	"context"
)

type ShotLooper interface {
	StartShootingLoop(ctx context.Context) (isWinner bool)
}
