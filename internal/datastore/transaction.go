package datastore

import (
	"errors"

	etcdClient "go.etcd.io/etcd/client/v3"
)

const OpTypePut = "put"

var ErrTransactionUnsuccessful = errors.New("transaction unsuccessful")

// Cmp represents a comparison in a transaction
type Cmp struct {
	key      string
	operator string
	value    string
}

// Op represents an operation in a transaction
type Op struct {
	opType string
	key    string
	value  string
}

type Transaction interface {
	If(...Cmp) *Txn
	Then(Op) *Txn
	Commit() error
}

type Txn struct {
	etcdTxn etcdClient.Txn
}

var _ Transaction = (*Txn)(nil)

// Compare creates a new Cmp instance
func Compare(key string, operator string, value string) Cmp {
	return Cmp{
		key:      key,
		operator: operator,
		value:    value,
	}
}

// OpPut creates a new Op instance for a put operation
func OpPut(key string, value string) Op {
	return Op{
		opType: OpTypePut,
		key:    key,
		value:  value,
	}
}

// If adds comparisons to the transaction and returns the updated transaction
func (t *Txn) If(cmps ...Cmp) *Txn {
	etcdCmps := make([]etcdClient.Cmp, 0, len(cmps))
	for _, cmp := range cmps {
		etcdCmps = append(etcdCmps, etcdClient.Compare(etcdClient.Value(cmp.key), cmp.operator, cmp.value))
	}

	t.etcdTxn = t.etcdTxn.If(etcdCmps...)

	return t
}

// Then adds an operation to the transaction and returns the updated transaction
func (t *Txn) Then(op Op) *Txn {
	if op.opType == OpTypePut {
		t.etcdTxn = t.etcdTxn.Then(etcdClient.OpPut(op.key, op.value))
	}

	return t
}

// Commit attempts to commit the transaction and returns an error if unsuccessful
func (t *Txn) Commit() error {
	resp, err := t.etcdTxn.Commit()
	if err != nil {
		return err
	}

	if !resp.Succeeded {
		return ErrTransactionUnsuccessful
	}

	return nil
}
