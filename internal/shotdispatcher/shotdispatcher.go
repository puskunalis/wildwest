package shotdispatcher

import "context"

type ShotDispatcher interface {
	Shoot(ctx context.Context, id int, from int64, damage int64) error
}
