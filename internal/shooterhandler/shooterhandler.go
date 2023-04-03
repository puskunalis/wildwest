package shooterhandler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"wildwest/internal/damagehandler"
	"wildwest/internal/datastore"
	"wildwest/internal/shotdispatcher"
	"wildwest/internal/utils"
)

var (
	ErrIAmTheWinner          = errors.New("i am the winner")
	ErrInvalidDatastoreState = errors.New("invalid datastore state")
)

type Shooter struct {
	logger     *zap.Logger
	id         int
	cowboy     *utils.Cowboy
	db         datastore.Datastore
	makeShot   chan struct{}
	shotSender *shotdispatcher.ShotDispatcher
}

func New(logger *zap.Logger, id int, cowboy *utils.Cowboy, db datastore.Datastore, shotSender *shotdispatcher.ShotDispatcher) *Shooter {
	return &Shooter{
		logger:     logger,
		id:         id,
		cowboy:     cowboy,
		db:         db,
		makeShot:   make(chan struct{}, 1),
		shotSender: shotSender,
	}
}

// startTicker starts a ticker that queues a shot every second
func (s *Shooter) startTicker(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go s.queueShot()
		}
	}
}

// StartShootingLoop begins shooting, exits once cowboy is either dead or the winner and returns true if cowboy won
func (s *Shooter) StartShootingLoop(ctx context.Context) bool {
	go s.startTicker(ctx)

	for {
		select {
		case <-ctx.Done():
			return false
		case <-s.makeShot:
			if err := s.shootAtRandomCowboy(); err != nil {
				if errors.Is(err, ErrIAmTheWinner) {
					return true
				}

				s.logger.Error("error shooting cowboy", zap.Error(err))
			}
		}
	}
}

func (s *Shooter) queueShot() {
	s.makeShot <- struct{}{}
}

// shootAtRandomCowboy finds a random alive cowboy and attempts to shoot him
func (s *Shooter) shootAtRandomCowboy() error {
	randomCowboyID, err := s.getRandomCowboyID()
	if err != nil {
		return err
	}

	// shoot the cowboy
	if err := s.shotSender.Shoot(randomCowboyID, s.cowboy.Damage, int64(s.id)); err != nil {
		// immediately retry if transaction was unsuccessful
		// TODO properly compare grpc errors
		if strings.HasSuffix(err.Error(), damagehandler.ErrVictimDead.Error()) ||
			strings.HasSuffix(err.Error(), damagehandler.ErrTransactionUnsuccessful.Error()) {
			go s.queueShot()
			return nil
		} else if strings.HasSuffix(err.Error(), damagehandler.ErrShooterDead.Error()) {
			return nil
		} else {
			return fmt.Errorf("send shot: %w", err)
		}
	}

	return nil
}

// getRandomCowboyID returns a random alive cowboy's id
func (s *Shooter) getRandomCowboyID() (int, error) {
	ourKey := utils.CowboyKeyPrefix + strconv.Itoa(s.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// get alive cowboy keys
	// TODO this will return unwanted keys if there are other keys with the prefix
	resp, err := s.db.GetPrefix(ctx, utils.CowboyKeyPrefix)
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
