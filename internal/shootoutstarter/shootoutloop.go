package shootoutstarter

import (
	"context"
	"strconv"
	"time"
	"wildwest/internal/datastore"
	"wildwest/internal/shotlooper"
	"wildwest/internal/utils"

	"go.uber.org/zap"
)

type Config struct {
	ID     int
	Cowboy utils.Cowboy

	DB              datastore.Datastore
	ShooterHandler  shotlooper.ShotLooper
	ShootoutManager ShootoutStarter

	Ready func()
}

func Start(ctx context.Context, logger *zap.Logger, cfg *Config) bool {
	dbCtx, dbCtxCancel := context.WithTimeout(ctx, time.Minute)
	defer dbCtxCancel()

	health, err := cfg.DB.Get(dbCtx, utils.CowboyKeyPrefix+strconv.Itoa(cfg.ID))
	// if we didn't find the health value already in the database
	if err != nil {
		logger.Debug("didn't find health already in the database")

		// initialize health in etcd
		err := cfg.DB.Put(dbCtx, utils.CowboyKeyPrefix+strconv.Itoa(cfg.ID), strconv.Itoa(int(cfg.Cowboy.Health)))
		if err != nil {
			logger.Fatal("set initial health value", zap.Error(err))
		}

		// start readiness server
		go cfg.Ready()

		logger.Info("waiting to begin shootout...")

		// wait until shootout beginning
		cfg.ShootoutManager.WaitForShootout()
		logger.Info("beginning shootout!")
	}

	if health == "0" {
		logger.Debug("found health already in the database, but we're already dead")
		return false
	}

	// if process restarted since health was found
	if err == nil {
		// start readiness server
		go cfg.Ready()
	}

	return cfg.ShooterHandler.StartShootingLoop(ctx)
}
