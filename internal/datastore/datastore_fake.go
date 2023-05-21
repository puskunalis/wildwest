package datastore

import (
	"context"
	"strings"
	"sync"
)

type FakeClient struct {
	kvStorage map[string]string
	mu        *sync.RWMutex
	dbMu      *sync.Mutex
}

var _ Datastore = (*FakeClient)(nil)

func (fc *FakeClient) Get(ctx context.Context, key string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	value, ok := fc.kvStorage[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	return value, nil
}

func (fc *FakeClient) GetPrefix(ctx context.Context, key string) (map[string]string, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

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

func (fc *FakeClient) Put(ctx context.Context, key string, value string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.kvStorage[key] = value

	return nil
}

func (fc *FakeClient) Transaction(ctx context.Context) Transaction {
	return &TxnFake{
		datastore: fc,
		ctx:       ctx,
	}
}

func (fc *FakeClient) Close() error {
	return nil
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		kvStorage: make(map[string]string),
		mu:        &sync.RWMutex{},
		dbMu:      &sync.Mutex{},
	}
}

func (fc *FakeClient) GetDBMu() *sync.Mutex {
	return fc.dbMu
}
