package damageapplier

import (
	"context"
)

type DamageApplier interface {
	ApplyDamage(ctx context.Context, from, damage int) (health int, err error)
	GetHealth(ctx context.Context) (health int, err error)
}
