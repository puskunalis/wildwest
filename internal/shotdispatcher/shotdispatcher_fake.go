package shotdispatcher

import (
	"context"
	"wildwest/internal/damageapplier"

	"go.uber.org/zap"
)

type FakeShotDispatcher struct {
	logger         *zap.Logger
	damageAppliers []damageapplier.DamageApplier
}

var _ ShotDispatcher = (*FakeShotDispatcher)(nil)

func NewFake(logger *zap.Logger, damageAppliers []damageapplier.DamageApplier) *FakeShotDispatcher {
	return &FakeShotDispatcher{
		logger:         logger,
		damageAppliers: damageAppliers,
	}
}

// Shoot sends the shot to another cowboy
func (fsd *FakeShotDispatcher) Shoot(ctx context.Context, id int, from int64, damage int64) error {
	_, err := fsd.damageAppliers[id].ApplyDamage(ctx, int(from), int(damage))
	return err
}
