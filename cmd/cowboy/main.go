package main

import (
	"context"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"wildwest/internal/damagehandler"
	"wildwest/internal/shooterhandler"
	"wildwest/internal/shootoutmanager"
	"wildwest/internal/shotdispatcher"
	"wildwest/internal/utils"

	"go.uber.org/zap"
)

const (
	podName            = "cowboy"
	serviceName        = "wildwest"
	namespace          = "default"
	replicasEnvKey     = "REPLICAS"
	etcdEndpoint       = "etcd.default.svc.cluster.local:2379"
	cowboyListFilename = "/wildwest/cowboys"
	grpcPort           = ":50051"
	readinessPort      = ":8080"
)

func main() {
	logger, syncLogger := utils.InitLogger()
	defer syncLogger() //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get our id
	id, err := utils.GetID()
	if err != nil {
		logger.Fatal("get id", zap.Error(err))
	}

	// add id to logger fields
	logger = logger.With(zap.Int("id", id))

	// get replica count
	replicas, err := utils.GetReplicas(replicasEnvKey)
	if err != nil {
		logger.Fatal("get replicas", zap.Error(err))
	}

	// get cowboys
	cowboys, err := utils.GetCowboys(cowboyListFilename, replicas)
	if err != nil {
		logger.Fatal("get cowboys", zap.Error(err))
	}

	// get our cowboy from cowboy list
	cowboy := &cowboys[id]

	// add cowboy name to logger fields
	logger = logger.With(zap.String("name", cowboy.Name))

	// init etcd
	db, err := utils.InitDatastore(etcdEndpoint)
	if err != nil {
		logger.Fatal("init datastore", zap.Error(err))
	}
	defer db.Close() //nolint:errcheck

	// listen on grpc port
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logger.Fatal("grpc server listen", zap.Error(err))
	}

	// define channel for receiving shootout start time
	shootoutTime := make(chan time.Time)

	// init grpc servers
	grpcServer := grpc.NewServer()
	cowboyServer := damagehandler.New(logger, cowboy, cancel, db, id)
	shootoutServer := shootoutmanager.New(shootoutTime)

	// register grpc servers
	cowboyServer.Register(grpcServer)
	shootoutServer.Register(grpcServer)

	// start grpc server
	go func(grpcServer *grpc.Server) {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("serve grpc server", zap.Error(err))
		}
	}(grpcServer)

	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dbCtxCancel()

	health, err := db.Get(dbCtx, utils.CowboyKeyPrefix+strconv.Itoa(id))
	// if we didn't find the health value already in the database
	if err != nil {
		logger.Debug("didn't find health already in the database")

		// initialize health in etcd
		err = db.Put(dbCtx, utils.CowboyKeyPrefix+strconv.Itoa(id), strconv.Itoa(int(cowboy.Health)))
		if err != nil {
			logger.Fatal("set initial health value", zap.Error(err))
		}

		// start readiness server
		go utils.StartReadinessServer(logger, readinessPort)

		logger.Info("waiting to begin shootout...")

		// wait until shootout beginning
		time.Sleep(time.Until(<-shootoutTime))
		logger.Info("beginning shootout!")
	} else {
		// start readiness server
		go utils.StartReadinessServer(logger, readinessPort)

		if health <= "0" {
			logger.Debug("found health already in the database, but we're already dead")
			select {}
		}
	}

	shotDispatcher := shotdispatcher.New(logger, podName, serviceName, namespace, grpcPort)

	shooterHandler := shooterhandler.New(logger, id, cowboy, db, shotDispatcher)

	isWinner := shooterHandler.StartShootingLoop(ctx)
	if isWinner {
		logger.Info("i am the winner!")
	}

	select {}
}
