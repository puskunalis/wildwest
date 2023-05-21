package targetprovider

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"wildwest/internal/datastore"
	"wildwest/internal/utils"
)

type DefaultTargetProvider struct {
	id int
	db datastore.Datastore
}

var _ TargetProvider = (*DefaultTargetProvider)(nil)

func New(id int, db datastore.Datastore) *DefaultTargetProvider {
	return &DefaultTargetProvider{
		id: id,
		db: db,
	}
}

// GetRandomTarget returns a random alive cowboy's id
func (dtp *DefaultTargetProvider) GetRandomTarget(ctx context.Context) (int, error) {
	ourKey := utils.CowboyKeyPrefix + strconv.Itoa(dtp.id)

	// get alive cowboy keys
	// TODO this will return unwanted keys if there are other keys with the prefix
	resp, err := dtp.db.GetPrefix(ctx, utils.CowboyKeyPrefix)
	if err != nil {
		return 0, fmt.Errorf("get alive cowboys: %w", err)
	}

	// filter out dead cowboys
	aliveCowboyKeys := make([]string, 0, len(resp))

	for k, v := range resp {
		if v > "0" {
			aliveCowboyKeys = append(aliveCowboyKeys, k)
		}
	}

	// if i am the only one left
	if len(aliveCowboyKeys) == 1 {
		if aliveCowboyKeys[0] == ourKey {
			return 0, ErrIAmTheWinner
		}
	}

	// if response is empty
	if len(aliveCowboyKeys) == 0 {
		return 0, ErrInvalidDatastoreState
	}

	// generate random idx
	randomIdx := rand.Intn(len(aliveCowboyKeys))
	for aliveCowboyKeys[randomIdx] == ourKey {
		randomIdx = rand.Intn(len(aliveCowboyKeys))
	}

	randomCowboyKey := aliveCowboyKeys[randomIdx]
	randomCowboyIDStr := strings.TrimPrefix(randomCowboyKey, utils.CowboyKeyPrefix)

	randomCowboyID, err := strconv.Atoi(randomCowboyIDStr)
	if err != nil {
		return 0, fmt.Errorf("convert random cowboy id to int: %w", err)
	}

	return randomCowboyID, nil
}
