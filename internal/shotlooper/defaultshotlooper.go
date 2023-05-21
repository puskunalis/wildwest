package shotlooper

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"wildwest/internal/datastore"
	"wildwest/internal/shotdispatcher"
	"wildwest/internal/shotqueue"
	"wildwest/internal/targetprovider"
	"wildwest/internal/utils"
)

type DefaultShotLooper struct {
	logger         *zap.Logger
	id             int
	cowboy         utils.Cowboy
	db             datastore.Datastore
	shotQueue      shotqueue.ShotQueue
	shotSender     shotdispatcher.ShotDispatcher
	targetProvider targetprovider.TargetProvider
}

var _ ShotLooper = (*DefaultShotLooper)(nil)

func New(logger *zap.Logger, id int, cowboy utils.Cowboy, db datastore.Datastore, shotQueue shotqueue.ShotQueue, shotSender shotdispatcher.ShotDispatcher, targetProvider targetprovider.TargetProvider) *DefaultShotLooper {
	return &DefaultShotLooper{
		logger:         logger,
		id:             id,
		cowboy:         cowboy,
		db:             db,
		shotQueue:      shotQueue,
		shotSender:     shotSender,
		targetProvider: targetProvider,
	}
}

// StartShootingLoop begins shooting, exits once cowboy is either dead or the winner and returns true if cowboy won
func (dsl *DefaultShotLooper) StartShootingLoop(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case <-dsl.shotQueue.DequeueShot():
			if err := dsl.shootAtRandomCowboy(ctx); err != nil {
				if errors.Is(err, targetprovider.ErrIAmTheWinner) {
					return true
				}

				if errors.Is(err, context.Canceled) {
					return false
				}

				dsl.logger.Error("error shooting cowboy", zap.Error(err))
			}
		}
	}
}

// shootAtRandomCowboy finds a random alive cowboy and attempts to shoot him
func (dsl *DefaultShotLooper) shootAtRandomCowboy(ctx context.Context) error {
	randomCowboyID, err := dsl.targetProvider.GetRandomTarget(ctx)
	if err != nil {
		return err
	}

	// shoot the cowboy
	if err := dsl.shotSender.Shoot(ctx, randomCowboyID, int64(dsl.id), dsl.cowboy.Damage); err != nil {
		// immediately retry if transaction was unsuccessful
		// TODO properly compare grpc errors
		if strings.HasSuffix(err.Error(), datastore.ErrTransactionUnsuccessful.Error()) {
			go dsl.shotQueue.QueueShot()
			return nil
		}

		return fmt.Errorf("send shot: %w", err)
	}

	return nil
}
