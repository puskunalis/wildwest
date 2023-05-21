package targetprovider

import (
	"context"
	"wildwest/internal/utils"
)

const (
	ErrIAmTheWinner          = utils.ConstError("i am the winner")
	ErrInvalidDatastoreState = utils.ConstError("invalid datastore state")
)

type TargetProvider interface {
	GetRandomTarget(ctx context.Context) (int, error)
}
