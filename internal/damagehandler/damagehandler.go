package damagehandler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"google.golang.org/protobuf/types/known/emptypb"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	damagepb "wildwest/api/proto/damage"
	"wildwest/internal/datastore"
	"wildwest/internal/utils"
)

var (
	ErrTransactionUnsuccessful = errors.New("transaction unsuccessful")
	ErrShooterDead             = errors.New("shooter is dead")
	ErrVictimDead              = errors.New("victim is dead")
)

type Server struct {
	damagepb.UnimplementedDamageServiceServer
	logger       *zap.Logger
	cowboy       *utils.Cowboy
	mu           *sync.Mutex
	observerFunc func()
	db           datastore.Datastore
	id           int
}

func New(logger *zap.Logger, cowboy *utils.Cowboy, observerFunc func(), db datastore.Datastore, id int) *Server {
	return &Server{
		logger:       logger,
		cowboy:       cowboy,
		mu:           &sync.Mutex{},
		observerFunc: observerFunc,
		db:           db,
		id:           id,
	}
}

// Register registers the damage handler server with the given gRPC service registrar
func (s *Server) Register(grpcServer grpc.ServiceRegistrar) {
	damagepb.RegisterDamageServiceServer(grpcServer, s)
}

// ApplyDamage handles applying damage to another cowboy
func (s *Server) ApplyDamage(ctx context.Context, req *damagepb.DamageRequest) (*emptypb.Empty, error) {
	logger := s.logger.With(
		zap.Int64("from", req.GetFrom()),
		zap.Int64("damage", req.GetDamage()),
	)

	logger.Debug("request received")

	shooterKey := utils.CowboyKeyPrefix + strconv.Itoa(int(req.GetFrom()))
	victimKey := utils.CowboyKeyPrefix + strconv.Itoa(s.id)

	s.mu.Lock()
	defer s.mu.Unlock()

	victimHealth, err := s.db.Get(ctx, victimKey)
	if err != nil {
		return &emptypb.Empty{}, fmt.Errorf("get victim health: %w", err)
	}

	if victimHealth <= "0" {
		return &emptypb.Empty{}, ErrVictimDead
	}

	myHealthStr := victimHealth

	myHealth, err := strconv.Atoi(myHealthStr)
	if err != nil {
		return &emptypb.Empty{}, fmt.Errorf("convert health to int: %w", err)
	}

	myHealth -= int(req.GetDamage())

	err = s.db.Transaction(ctx).If(
		datastore.Compare(shooterKey, ">", "0"),
		datastore.Compare(victimKey, ">", "0"),
	).Then(
		datastore.OpPut(victimKey, strconv.Itoa(myHealth)),
	).Commit()
	if err != nil {
		return &emptypb.Empty{}, fmt.Errorf("commit transaction: %w", err)
	}

	logger = logger.With(zap.Int("health", myHealth))

	if myHealth <= 0 {
		// execute observer function
		s.observerFunc()
		logger.Info("killing shot received")
	} else {
		logger.Info("shot received")
	}

	return &emptypb.Empty{}, nil
}
