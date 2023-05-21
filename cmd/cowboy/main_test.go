package main

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"wildwest/internal/damageapplier"
	"wildwest/internal/datastore"
	"wildwest/internal/shootoutstarter"
	"wildwest/internal/shotqueue"
	"wildwest/internal/targetprovider"

	"github.com/stretchr/testify/assert"
	"wildwest/internal/shotdispatcher"
	"wildwest/internal/shotlooper"
	"wildwest/internal/utils"

	"go.uber.org/zap"
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateRandomString(r *rand.Rand, n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}

	return string(b)
}

func TestMainFunc(t *testing.T) {
	replicas := 1000
	shotFrequencyMs := 1

	// init datastore
	db := datastore.NewFakeClient()

	damageAppliers := make([]damageapplier.DamageApplier, replicas)

	shootoutBeginTime := time.Now().Add(2 * time.Second).Round(time.Second)

	winnersCount := atomic.Uint64{}
	deadCount := atomic.Uint64{}

	r := rand.New(rand.NewSource(int64(replicas)))

	// generate pseudorandom list of cowboys
	cowboys := make([]utils.Cowboy, 0, replicas)
	for i := 0; i < replicas; i++ {
		cowboys = append(cowboys, utils.Cowboy{
			Name:   generateRandomString(r, 32),
			Health: 1 + int64(r.Intn(100)),
			Damage: 1 + int64(r.Intn(50)),
		})
	}

	wg := &sync.WaitGroup{}
	wg.Add(replicas)
	for i := 0; i < replicas; i++ {
		go func(id int) {
			logger := zap.NewNop()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			killingShotReceivedFunc := func() {
				cancel()
				deadCount.Add(1)
			}

			// add id to logger fields
			logger = logger.With(zap.Int("id", id))

			// get our cowboy from cowboy list
			cowboy := cowboys[id]

			// add name to logger fields
			logger = logger.With(zap.String("name", cowboy.Name))

			// init damage applier
			damageAppliers[id] = damageapplier.New(logger, id, db, killingShotReceivedFunc)

			shotQueue := shotqueue.New(time.Duration(shotFrequencyMs) * time.Millisecond)
			shotDispatcher := shotdispatcher.NewFake(logger, damageAppliers)
			targetProvider := targetprovider.New(id, db)

			shooterHandler := shotlooper.New(logger, id, cowboy, db, shotQueue, shotDispatcher, targetProvider)

			shootoutManager := shootoutstarter.New()

			// mock call to begin shootout
			shootoutManager.ReceiveShootoutTime(shootoutBeginTime)

			isWinner := shootoutstarter.Start(ctx, logger, &shootoutstarter.Config{
				ID:              id,
				Cowboy:          cowboy,
				DB:              db,
				ShooterHandler:  shooterHandler,
				ShootoutManager: shootoutManager,
				Ready:           func() {},
			})

			if isWinner {
				winnersCount.Add(1)
				logger.Info("i am the winner!")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	assert.Equal(t, uint64(1), winnersCount.Load())
	assert.Equal(t, uint64(replicas-1), deadCount.Load())
}
