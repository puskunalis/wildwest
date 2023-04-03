package shootoutmanager

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc"

	shootoutpb "wildwest/api/proto/shootout"
)

type Server struct {
	shootoutpb.UnimplementedShootoutServiceServer
	shootoutTime chan<- time.Time
}

func New(shootoutTime chan<- time.Time) *Server {
	return &Server{
		shootoutTime: shootoutTime,
	}
}

// Register registers the damage handler server with the given gRPC service registrar
func (s *Server) Register(grpcServer grpc.ServiceRegistrar) {
	shootoutpb.RegisterShootoutServiceServer(grpcServer, s)
}

// BeginShootout sends the shootout beginning time to the held channel
func (s *Server) BeginShootout(_ context.Context, req *shootoutpb.BeginShootoutRequest) (*emptypb.Empty, error) {
	s.shootoutTime <- time.Unix(req.Timestamp, 0)

	return &emptypb.Empty{}, nil
}
