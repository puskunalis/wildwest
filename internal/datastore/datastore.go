package datastore

import (
	"context"
	"errors"

	etcdClient "go.etcd.io/etcd/client/v3"
)

var ErrKeyNotFound = errors.New("key not found")

type Datastore interface {
	Get(ctx context.Context, key string) (string, error)
	GetPrefix(ctx context.Context, key string) (map[string]string, error)
	Put(ctx context.Context, key string, val string) error
	Transaction(ctx context.Context) *Txn
	Close() error
}

type KV struct {
	Key   string
	Value string
}

// GetResponse represents the result of a Get operation
type GetResponse KV

// GetPrefixResponse represents the result of a GetPrefix operation
type GetPrefixResponse []KV

type EtcdClientWrapper struct {
	client              etcdClient.KV
	closeConnectionFunc func() error
}

var _ Datastore = (*EtcdClientWrapper)(nil)

func New(client *etcdClient.Client) *EtcdClientWrapper {
	return &EtcdClientWrapper{
		client:              etcdClient.NewKV(client),
		closeConnectionFunc: client.Close,
	}
}

// Get retrieves the value associated with the given key, or returns an error if the key is not found
func (ecw *EtcdClientWrapper) Get(ctx context.Context, key string) (string, error) {
	resp, err := ecw.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if resp.Count == 0 {
		return "", ErrKeyNotFound
	}

	return string(resp.Kvs[0].Value), nil
}

// GetPrefix retrieves a map of key-value pairs with keys that have the given prefix
func (ecw *EtcdClientWrapper) GetPrefix(ctx context.Context, key string) (map[string]string, error) {
	resp, err := ecw.client.Get(ctx, key, etcdClient.WithPrefix())
	if err != nil {
		return nil, err
	}

	getPrefixResponse := make(map[string]string)
	for _, kv := range resp.Kvs {
		getPrefixResponse[string(kv.Key)] = string(kv.Value)
	}

	return getPrefixResponse, nil
}

// Put stores the given key-value pair
func (ecw *EtcdClientWrapper) Put(ctx context.Context, key string, value string) error {
	_, err := ecw.client.Put(ctx, key, value)

	return err
}

// Transaction creates a new transaction
func (ecw *EtcdClientWrapper) Transaction(ctx context.Context) *Txn {
	return &Txn{
		etcdTxn: ecw.client.Txn(ctx),
	}
}

// Close closes the connection to the datastore
func (ecw *EtcdClientWrapper) Close() error {
	return ecw.closeConnectionFunc()
}
