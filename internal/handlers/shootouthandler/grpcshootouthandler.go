package shootouthandler

import (
	"context"
	"time"
	shootoutpb "wildwest/api/proto/shootout"
	"wildwest/internal/shootoutstarter"

	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCShootoutHandler struct {
	shootoutpb.UnimplementedShootoutServiceServer
	shootoutManager shootoutstarter.ShootoutStarter
}

func NewGRPC(shootoutManager shootoutstarter.ShootoutStarter) *GRPCShootoutHandler {
	return &GRPCShootoutHandler{
		shootoutManager: shootoutManager,
	}
}

// ReceiveShootoutTime receives the shootout beginning time and sends it to the held channel
func (sh *GRPCShootoutHandler) ReceiveShootoutTime(_ context.Context, req *shootoutpb.ReceiveShootoutTimeRequest) (*emptypb.Empty, error) {
	sh.shootoutManager.ReceiveShootoutTime(time.Unix(req.Timestamp, 0))
	return &emptypb.Empty{}, nil
}
