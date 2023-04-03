package main

import (
	"wildwest/internal/broadcastdispatcher"
	"wildwest/internal/utils"

	"go.uber.org/zap"
)

const (
	podName        = "cowboy"
	serviceName    = "wildwest"
	namespace      = "default"
	replicasEnvKey = "REPLICAS"
	grpcPort       = ":50051"
	readinessPort  = ":8080"
)

func main() {
	logger, syncLogger := utils.InitLogger()
	defer syncLogger() //nolint:errcheck

	// start readiness server
	go utils.StartReadinessServer(logger, readinessPort)

	// get replica count
	replicas, err := utils.GetReplicas(replicasEnvKey)
	if err != nil {
		logger.Fatal("get replicas", zap.Error(err))
	}

	broadcastdispatcher.BroadcastShootoutTime(logger, replicas, podName, serviceName, namespace, grpcPort)

	select {}
}
