package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"
	damagepb "wildwest/api/proto/damage"
	shootoutpb "wildwest/api/proto/shootout"
	"wildwest/internal/damageapplier"
	"wildwest/internal/datastore"
	"wildwest/internal/handlers/damagehandler"
	"wildwest/internal/handlers/shootouthandler"
	"wildwest/internal/shotqueue"
	"wildwest/internal/targetprovider"

	"github.com/caarlos0/env/v6"

	"google.golang.org/grpc"

	"wildwest/internal/shootoutstarter"
	"wildwest/internal/shotdispatcher"
	"wildwest/internal/shotlooper"
	"wildwest/internal/utils"

	"go.uber.org/zap"
)

// TODO graceful shutdown

func main() {
	logger := utils.InitLogger()
	defer logger.Sync() //nolint:errcheck

	// parse environment variables
	var envConfig utils.Environment
	err := env.Parse(&envConfig)
	if err != nil {
		logger.Fatal("parse environment", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get our id
	id, err := utils.GetID(os.Hostname)
	if err != nil {
		logger.Fatal("get id", zap.Error(err))
	}

	// add id to logger fields
	logger = logger.With(zap.Int("id", id))

	// get cowboys
	cowboys, err := utils.GetCowboys(envConfig.CowboyListFilePath, envConfig.Replicas)
	if err != nil {
		logger.Fatal("get cowboys", zap.Error(err))
	}

	// get our cowboy from cowboy list
	cowboy := cowboys[id]

	// add name to logger fields
	logger = logger.With(zap.String("name", cowboy.Name))

	// init etcd
	db, err := datastore.InitEtcdDatastore(fmt.Sprintf("%s:%d", envConfig.EtcdAppName, envConfig.EtcdPort))
	if err != nil {
		logger.Fatal("init datastore", zap.Error(err))
	}
	defer db.Close() //nolint:errcheck

	// listen on grpc port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", envConfig.GRPCPort))
	if err != nil {
		logger.Fatal("grpc server listen", zap.Error(err))
	}

	// init damage applier
	damageApplier := damageapplier.New(logger, id, db, cancel)

	// init shootout manager
	shootoutManager := shootoutstarter.New()

	// init grpc servers
	grpcServer := grpc.NewServer()

	damageHandler := damagehandler.NewGRPC(logger, damageApplier)
	damagepb.RegisterDamageServiceServer(grpcServer, damageHandler)

	shootoutHandler := shootouthandler.NewGRPC(shootoutManager)
	shootoutpb.RegisterShootoutServiceServer(grpcServer, shootoutHandler)

	// start grpc server
	go func(grpcServer *grpc.Server) {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("serve grpc server", zap.Error(err))
		}
	}(grpcServer)

	shotQueue := shotqueue.New(time.Duration(envConfig.ShotFreqMs) * time.Millisecond)
	shotDispatcher := shotdispatcher.NewGRPC(logger, envConfig.CowboyAppName, envConfig.CowboyAppName, envConfig.GRPCPort)
	targetProvider := targetprovider.New(id, db)

	shooterHandler := shotlooper.New(logger, id, cowboy, db, shotQueue, shotDispatcher, targetProvider)

	isWinner := shootoutstarter.Start(ctx, logger, &shootoutstarter.Config{
		ID:              id,
		Cowboy:          cowboy,
		DB:              db,
		ShooterHandler:  shooterHandler,
		ShootoutManager: shootoutManager,
		Ready:           func() { utils.StartReadinessServer(logger, envConfig.ReadinessPort) },
	})

	if isWinner {
		logger.Info("i am the winner!")
	}

	select {}
}
