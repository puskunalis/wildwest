package damagehandler

import (
	"context"
	"wildwest/internal/damageapplier"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	damagepb "wildwest/api/proto/damage"
)

type GRPCDamageHandler struct {
	damagepb.UnimplementedDamageServiceServer
	logger        *zap.Logger
	damageApplier damageapplier.DamageApplier
}

func NewGRPC(logger *zap.Logger, damageApplier damageapplier.DamageApplier) *GRPCDamageHandler {
	return &GRPCDamageHandler{
		logger:        logger,
		damageApplier: damageApplier,
	}
}

func (dh *GRPCDamageHandler) ReceiveDamage(ctx context.Context, req *damagepb.DamageRequest) (*emptypb.Empty, error) {
	_, err := dh.damageApplier.ApplyDamage(ctx, int(req.GetFrom()), int(req.GetDamage()))
	return &emptypb.Empty{}, err
}
