package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const CowboyKeyPrefix = "cowboy-"

type Cowboy struct {
	Name   string `json:"name"`
	Health int64  `json:"health"`
	Damage int64  `json:"damage"`
}

type Environment struct {
	CowboyListFilePath string `env:"COWBOY_LIST_FILE_PATH"`
	ShotFreqMs         int    `env:"SHOT_FREQ_MS"`
	Replicas           int    `env:"REPLICAS"`
	CowboyAppName      string `env:"COWBOY_APP_NAME"`
	EtcdAppName        string `env:"ETCD_APP_NAME"`
	GRPCPort           int    `env:"GRPC_PORT"`
	ReadinessPort      int    `env:"READINESS_PORT"`
	EtcdPort           int    `env:"ETCD_PORT"`
}

func InitLogger() *zap.Logger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logger, _ := loggerConfig.Build()

	return logger
}
