package datastore

import (
	"context"
)

type TxnFake struct {
	datastore *FakeClient
	cmps      []Cmp
	op        Op

	ctx context.Context
}

var _ Transaction = (*TxnFake)(nil)

// If adds comparisons to the transaction and returns the updated transaction
func (tf *TxnFake) If(cmps ...Cmp) Transaction {
	tf.cmps = cmps

	return tf
}

// Then adds an operation to the transaction and returns the updated transaction
func (tf *TxnFake) Then(op Op) Transaction {
	tf.op = op

	return tf
}

// Commit attempts to commit the transaction and returns an error if unsuccessful
func (tf *TxnFake) Commit() error {
	if tf.ctx.Err() != nil {
		return tf.ctx.Err()
	}

	tf.datastore.GetDBMu().Lock()
	defer tf.datastore.GetDBMu().Unlock()

	for _, cmp := range tf.cmps {
		if val, _ := tf.datastore.Get(context.Background(), cmp.key); val <= cmp.value {
			return ErrTransactionUnsuccessful
		}
	}

	return tf.datastore.Put(context.Background(), tf.op.key, tf.op.value)
}
