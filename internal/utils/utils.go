package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	etcdClient "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"wildwest/internal/datastore"
)

const (
	CowboyKeyPrefix = "cowboy-"
)

var (
	ErrInvalidHostnameFormat = errors.New("invalid hostname format")
	ErrEnvVariableNotFound   = errors.New("given environment variable not found")
	ErrNotEnoughReplicas     = errors.New("must have at least 2 replicas")
)

type Cowboy struct {
	Name   string `json:"name"`
	Health int64  `json:"health"`
	Damage int64  `json:"damage"`
}

func InitLogger() (*zap.Logger, func() error) {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logger, _ := loggerConfig.Build()

	return logger, logger.Sync
}

func InitDatastore(etcdEndpoint string) (*datastore.EtcdClientWrapper, error) {
	etcdEndpoints := []string{etcdEndpoint}
	cfg := etcdClient.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: time.Minute,
	}

	etcdC, err := etcdClient.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("create etcd client: %w", err)
	}

	return datastore.New(etcdC), nil
}

// GetReplicas returns the replica count from a given environment variable
func GetReplicas(envKey string) (int, error) {
	// get replicas count string from an environment variable
	replicasStr, ok := os.LookupEnv(envKey)
	if !ok {
		return 0, ErrEnvVariableNotFound
	}

	// convert replicas count string to int
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil {
		return 0, fmt.Errorf("convert '%s' environment variable (value: '%s') to int: %w", envKey, replicasStr, err)
	}

	// check if we have enough replicas
	if replicas <= 1 {
		return 0, ErrNotEnoughReplicas
	}

	return replicas, nil
}

// GetID gets our pod ID from the StatefulSet hostname format (pod-0, pod-1, pod-2...)
func GetID() (int, error) {
	// get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return 0, fmt.Errorf("get hostname: %w", err)
	}

	// get id from hostname
	split := strings.Split(hostname, "-")
	if len(split) != 2 {
		return 0, ErrInvalidHostnameFormat
	}

	// retrieve id
	idStr := split[1]

	// convert id to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("convert id string to int: %w", err)
	}

	return id, nil
}

// StartReadinessServer starts a server on :8080/ready, call is blocking
func StartReadinessServer(logger *zap.Logger, addr string) {
	http.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal("readiness server listen", zap.Error(err))
	}
}
