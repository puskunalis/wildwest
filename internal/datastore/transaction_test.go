package datastore_test

import (
	"context"
	"testing"
	"wildwest/internal/datastore"

	"github.com/stretchr/testify/assert"
)

const opKey = "opKey"

func TestTransactions(t *testing.T) {
	tests := []struct {
		name        string
		setupValues map[string]string
		cmps        []datastore.Cmp
		op          datastore.Op
		wantErr     error
		want        string
	}{
		{
			name: "successful transaction",
			setupValues: map[string]string{
				"1":   "1",
				opKey: "1",
			},
			cmps: []datastore.Cmp{
				datastore.Compare("1", ">", "0"),
				datastore.Compare(opKey, ">", "0"),
			},
			op: datastore.OpPut(opKey, "0"),

			wantErr: nil,
			want:    "0",
		},
		{
			name: "unsuccessful transaction",
			setupValues: map[string]string{
				"1":   "0",
				opKey: "1",
			},
			cmps: []datastore.Cmp{
				datastore.Compare("1", ">", "0"),
				datastore.Compare(opKey, ">", "0"),
			},
			op:      datastore.OpPut(opKey, "0"),
			wantErr: datastore.ErrTransactionUnsuccessful,
			want:    "1",
		},
		{
			name: "unsuccessful transaction checking non-existent key",
			setupValues: map[string]string{
				opKey: "1",
			},
			cmps: []datastore.Cmp{
				datastore.Compare("1", ">", "0"),
				datastore.Compare(opKey, ">", "0"),
			},
			op:      datastore.OpPut(opKey, "0"),
			wantErr: datastore.ErrTransactionUnsuccessful,
			want:    "1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			fakeDatastore := datastore.NewFakeClient()

			ctx := context.Background()

			for k, v := range tc.setupValues {
				err := fakeDatastore.Put(ctx, k, v)
				assert.NoError(t, err)
			}

			// execute
			err := fakeDatastore.Transaction(ctx).If(tc.cmps...).Then(tc.op).Commit()

			// verify
			assert.ErrorIs(t, err, tc.wantErr)

			got, err := fakeDatastore.Get(ctx, opKey)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
