package damageapplier

import (
	"context"
	"strconv"
	"sync"
	"wildwest/internal/datastore"
	"wildwest/internal/utils"

	"go.uber.org/zap"
)

type DefaultDamageApplier struct {
	logger       *zap.Logger
	id           int
	db           datastore.Datastore
	mu           *sync.RWMutex
	observerFunc func()
}

var _ DamageApplier = (*DefaultDamageApplier)(nil)

func New(logger *zap.Logger, id int, db datastore.Datastore, observerFunc func()) *DefaultDamageApplier {
	return &DefaultDamageApplier{
		logger:       logger,
		id:           id,
		db:           db,
		mu:           &sync.RWMutex{},
		observerFunc: observerFunc,
	}
}

func (da *DefaultDamageApplier) ApplyDamage(ctx context.Context, from, damage int) (int, error) {
	da.mu.Lock()
	defer da.mu.Unlock()

	logger := da.logger.With(
		zap.Int("from", from),
		zap.Int("damage", damage),
	)

	receiverHealth, err := da.getHealthNoLock(ctx)
	if err != nil {
		return 0, err
	}

	newReceiverHealth := receiverHealth - damage
	if newReceiverHealth < 0 {
		newReceiverHealth = 0
	}

	err = da.db.Transaction(ctx).If(
		datastore.Compare(utils.CowboyKeyPrefix+strconv.Itoa(da.id), ">", "0"),
		datastore.Compare(utils.CowboyKeyPrefix+strconv.Itoa(from), ">", "0"),
	).Then(
		datastore.OpPut(utils.CowboyKeyPrefix+strconv.Itoa(da.id), strconv.Itoa(newReceiverHealth)),
	).Commit()
	if err != nil {
		return 0, err
	}

	logger = logger.With(zap.Int("health", newReceiverHealth))

	if newReceiverHealth <= 0 {
		// execute observer function
		da.observerFunc()
		logger.Info("killing shot received")
		return 0, nil
	}

	logger.Info("shot received")

	return newReceiverHealth, nil
}

func (da *DefaultDamageApplier) getHealthNoLock(ctx context.Context) (int, error) {
	receiverHealthStr, err := da.db.Get(ctx, utils.CowboyKeyPrefix+strconv.Itoa(da.id))
	if err != nil {
		return 0, err
	}

	receiverHealth, err := strconv.Atoi(receiverHealthStr)
	if err != nil {
		return 0, err
	}

	return receiverHealth, nil
}

func (da *DefaultDamageApplier) GetHealth(ctx context.Context) (int, error) {
	da.mu.RLock()
	defer da.mu.RUnlock()

	return da.getHealthNoLock(ctx)
}
