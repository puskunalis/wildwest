package datastore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestMockClient(t *testing.T) {
	fakeClient := newFakeClient()

	tests := []struct {
		name           string
		key            string
		value          string
		getPrefix      string
		getPrefixError error
		getPrefixRes   map[string]string
	}{
		{"test key", "test123", "val", "key", ErrKeyNotFound, nil},
		{"key 1", "key1", "value1", "key", nil, map[string]string{"key1": "value1"}},
		{"key 2", "key2", "value2", "key", nil, map[string]string{"key1": "value1", "key2": "value2"}},
		{"key 3", "key3", "value3", "key", nil, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}},
		{"foo key", "foo", "bar", "foo", nil, map[string]string{"foo": "bar"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := fakeClient.Put(context.Background(), tc.key, tc.value)
			assert.NoError(t, err)

			value, err := fakeClient.Get(context.Background(), tc.key)
			assert.NoError(t, err)
			assert.Equal(t, tc.value, value)

			prefixValues, err := fakeClient.GetPrefix(context.Background(), tc.getPrefix)
			assert.Equal(t, tc.getPrefixError, err)

			if diff := cmp.Diff(tc.getPrefixRes, prefixValues); diff != "" {
				t.Errorf(diff)
			}
		})
	}

	t.Run("key not found", func(t *testing.T) {
		value, err := fakeClient.Get(context.Background(), "nonexistent")
		assert.Equal(t, ErrKeyNotFound, err)
		assert.Empty(t, value)
	})
}
