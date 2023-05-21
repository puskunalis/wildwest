package datastore_test

import (
	"context"
	"testing"
	"wildwest/internal/datastore"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
		err      error
	}{
		{
			name:     "Existing key",
			key:      "key1",
			value:    "value1",
			expected: "value1",
			err:      nil,
		},
		{
			name:     "Non-existing key",
			key:      "key2",
			value:    "",
			expected: "",
			err:      datastore.ErrKeyNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := datastore.NewFakeClient()
			ctx := context.Background()

			if tc.value != "" {
				err := client.Put(ctx, tc.key, tc.value)
				assert.NoError(t, err)
			}
			value, err := client.Get(ctx, tc.key)
			assert.Equal(t, tc.expected, value)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestGetPrefix(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		kvPairs  map[string]string
		expected map[string]string
		err      error
	}{
		{
			name: "Multiple matching keys",
			key:  "key",
			kvPairs: map[string]string{
				"key1":   "value1",
				"key2":   "value2",
				"random": "random",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			err: nil,
		},
		{
			name:     "No matching keys",
			key:      "key",
			kvPairs:  map[string]string{"random": "random"},
			expected: nil,
			err:      datastore.ErrKeyNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := datastore.NewFakeClient()
			ctx := context.Background()

			for key, value := range tc.kvPairs {
				err := client.Put(ctx, key, value)
				assert.NoError(t, err)
			}
			value, err := client.GetPrefix(ctx, tc.key)
			assert.Equal(t, tc.expected, value)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestPut(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "Put value",
			key:   "key1",
			value: "value1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := datastore.NewFakeClient()
			ctx := context.Background()

			err := client.Put(ctx, tc.key, tc.value)
			assert.NoError(t, err)
			val, _ := client.Get(ctx, tc.key)
			assert.Equal(t, tc.value, val)
		})
	}
}
