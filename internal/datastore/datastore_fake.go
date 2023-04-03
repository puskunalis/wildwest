package datastore

import (
	"context"
	"strings"
)

type fakeClient struct {
	kvStorage map[string]string
}

var _ Datastore = (*fakeClient)(nil)

func (fc *fakeClient) Get(_ context.Context, key string) (string, error) {
	value, ok := fc.kvStorage[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	return value, nil
}

func (fc *fakeClient) GetPrefix(_ context.Context, key string) (map[string]string, error) {
	resp := make(map[string]string)

	for k, v := range fc.kvStorage {
		if strings.HasPrefix(k, key) {
			resp[k] = v
		}
	}

	if len(resp) == 0 {
		return nil, ErrKeyNotFound
	}

	return resp, nil
}

func (fc *fakeClient) Put(_ context.Context, key string, value string) error {
	fc.kvStorage[key] = value

	return nil
}

func (fc *fakeClient) Transaction(_ context.Context) *Txn {
	return nil
}

func (fc *fakeClient) Close() error {
	return nil
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		kvStorage: make(map[string]string),
	}
}
