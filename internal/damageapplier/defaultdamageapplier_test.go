package damageapplier_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"wildwest/internal/damageapplier"
	"wildwest/internal/datastore"
	"wildwest/internal/utils"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestApplyDamage(t *testing.T) {
	type action struct {
		shooterID     int
		shooterHealth int
		shooterDamage int
	}

	tests := []struct {
		name                string
		receiverID          int
		receiverStartHealth int
		receiverEndHealth   int
		actions             []action
	}{
		{"shots below 0 health", 1, 20, 0, []action{
			{2, 5, 2},
			{2, 5, 2},
			{2, 5, 2},
			{3, 0, 2},
			{4, 0, 10},
			{5, 1, 5},
			{5, 1, 5},
			{6, 3, 8},
		}},
		{"shots exactly to 0 health", 1, 20, 0, []action{
			{2, 5, 5},
			{2, 5, 5},
			{2, 5, 5},
			{2, 5, 5},
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			fakeDatastore := datastore.NewFakeClient()

			err := fakeDatastore.Put(context.Background(), utils.CowboyKeyPrefix+strconv.Itoa(tc.receiverID), strconv.Itoa(tc.receiverStartHealth))
			assert.NoError(t, err)

			damageReceiver := damageapplier.New(zap.NewNop(), tc.receiverID, fakeDatastore, func() {})

			for _, a := range tc.actions {
				err := fakeDatastore.Put(context.Background(), utils.CowboyKeyPrefix+strconv.Itoa(a.shooterID), strconv.Itoa(a.shooterHealth))
				assert.NoError(t, err)
			}

			// execute
			for _, a := range tc.actions {
				_, _ = damageReceiver.ApplyDamage(context.Background(), a.shooterID, a.shooterDamage)
			}

			// verify
			gotReceiverEndHealth, err := fakeDatastore.Get(context.Background(), utils.CowboyKeyPrefix+strconv.Itoa(tc.receiverID))
			assert.NoError(t, err)
			assert.Equal(t, strconv.Itoa(tc.receiverEndHealth), gotReceiverEndHealth)
		})
	}
}

func TestApplyDamageParallel(t *testing.T) {
	tests := []struct {
		name        string
		startHealth int
		shots       int
		endHealth   int
	}{
		{"100 shots", 10000, 100, 4950},
		{"1000 shots", 1000000, 1000, 499500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			fakeDatastore := datastore.NewFakeClient()

			err := fakeDatastore.Put(context.Background(), utils.CowboyKeyPrefix+"1", strconv.Itoa(tc.startHealth))
			assert.NoError(t, err)

			err = fakeDatastore.Put(context.Background(), utils.CowboyKeyPrefix+"2", "1")
			assert.NoError(t, err)

			damageReceiver := damageapplier.New(zap.NewNop(), 1, fakeDatastore, func() {})

			// execute
			wg := sync.WaitGroup{}
			wg.Add(tc.shots)

			for i := 1; i <= tc.shots; i++ {
				go func(damage int) {
					defer wg.Done()
					_, _ = damageReceiver.ApplyDamage(context.Background(), 2, damage)
				}(i)
			}

			wg.Wait()

			// verify
			health, err := damageReceiver.GetHealth(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tc.endHealth, health)
		})
	}
}
