package shotdispatcher

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	damagepb "wildwest/api/proto/damage"
)

type CowboyClient struct {
	client damagepb.DamageServiceClient
}

type ShotDispatcher struct {
	logger        *zap.Logger
	podName       string
	serviceName   string
	namespace     string
	grpcPort      string
	cowboyClients map[int]*CowboyClient
}

func New(logger *zap.Logger, podName string, serviceName string, namespace string, grpcPort string) *ShotDispatcher {
	return &ShotDispatcher{
		logger:        logger,
		podName:       podName,
		serviceName:   serviceName,
		namespace:     namespace,
		grpcPort:      grpcPort,
		cowboyClients: make(map[int]*CowboyClient),
	}
}

// createCowboyclient establishes a connection to another cowboy and returns a client
func createCowboyClient(logger *zap.Logger, podName string, serviceName string, namespace string, grpcPort string, id int) *CowboyClient {
	hostname := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local%s", podName, id, serviceName, namespace, grpcPort)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// dial cowboy
	conn, err := grpc.DialContext(ctx, hostname, grpc.WithTransportCredentials(insecure.NewCredentials())) // TODO insecure
	if err != nil {
		logger.Fatal("failed to dial", zap.Error(err)) // TODO propagate?
	}

	// create client
	client := damagepb.NewDamageServiceClient(conn)

	return &CowboyClient{
		client: client,
	}
}

// getClient returns the CowboyClient lazily (creates it if it hasn't been created yet)
func (sd *ShotDispatcher) getClient(id int) *CowboyClient {
	cowboyClient, ok := sd.cowboyClients[id]
	if !ok {
		newClient := createCowboyClient(sd.logger, sd.podName, sd.serviceName, sd.namespace, sd.grpcPort, id)
		sd.cowboyClients[id] = newClient

		return newClient
	}

	return cowboyClient
}

// Shoot sends the shot to another cowboy
func (sd *ShotDispatcher) Shoot(id int, damage int64, from int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c := sd.getClient(id)

	// apply damage
	_, err := c.client.ApplyDamage(ctx, &damagepb.DamageRequest{Damage: damage, From: from})
	if err != nil {
		return fmt.Errorf("failed to send damage: %w", err)
	}

	return nil
}
