package main

import (
	"wildwest/internal/broadcastdispatcher"
	"wildwest/internal/utils"

	"github.com/caarlos0/env/v6"

	"go.uber.org/zap"
)

func main() {
	logger := utils.InitLogger()
	defer logger.Sync() //nolint:errcheck

	// parse environment variables
	var envConfig utils.Environment
	err := env.Parse(&envConfig)
	if err != nil {
		logger.Fatal("parse environment", zap.Error(err))
	}

	// start readiness server
	go utils.StartReadinessServer(logger, envConfig.ReadinessPort)

	broadcastdispatcher.BroadcastShootoutTime(logger,
		envConfig.Replicas,
		envConfig.CowboyAppName,
		envConfig.CowboyAppName,
		envConfig.GRPCPort)

	select {}
}
