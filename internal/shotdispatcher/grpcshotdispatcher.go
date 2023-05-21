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

type GRPCShotDispatcher struct {
	logger        *zap.Logger
	podName       string
	serviceName   string
	grpcPort      int
	cowboyClients map[int]*CowboyClient
}

var _ ShotDispatcher = (*GRPCShotDispatcher)(nil)

func NewGRPC(logger *zap.Logger, podName string, serviceName string, grpcPort int) *GRPCShotDispatcher {
	return &GRPCShotDispatcher{
		logger:        logger,
		podName:       podName,
		serviceName:   serviceName,
		grpcPort:      grpcPort,
		cowboyClients: make(map[int]*CowboyClient),
	}
}

// createCowboyclient establishes a connection to another cowboy and returns a client
func createCowboyClient(podName string, serviceName string, grpcPort int, id int) (*CowboyClient, error) {
	hostname := fmt.Sprintf("%s-%d.%s:%d", podName, id, serviceName, grpcPort)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// dial cowboy
	conn, err := grpc.DialContext(ctx, hostname, grpc.WithTransportCredentials(insecure.NewCredentials())) // TODO insecure
	if err != nil {
		return nil, err
	}

	// create client
	client := damagepb.NewDamageServiceClient(conn)

	return &CowboyClient{
		client: client,
	}, nil
}

// getClient returns the CowboyClient lazily (creates it if it hasn't been created yet)
func (gsd *GRPCShotDispatcher) getClient(id int) (*CowboyClient, error) {
	// if cowboy client was already created
	if cowboyClient, ok := gsd.cowboyClients[id]; ok {
		return cowboyClient, nil
	}

	// create new cowboy client and store it in map
	newClient, err := createCowboyClient(gsd.podName, gsd.serviceName, gsd.grpcPort, id)
	if err != nil {
		return nil, err
	}

	gsd.cowboyClients[id] = newClient

	return newClient, nil
}

// Shoot sends the shot to another cowboy
func (gsd *GRPCShotDispatcher) Shoot(ctx context.Context, id int, from int64, damage int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	c, err := gsd.getClient(id)
	if err != nil {
		return err
	}

	// apply damage
	_, err = c.client.ReceiveDamage(ctx, &damagepb.DamageRequest{From: from, Damage: damage})
	if err != nil {
		return fmt.Errorf("failed to send damage: %w", err)
	}

	return nil
}
