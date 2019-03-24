package v2

import (
	"github.com/jmoiron/sqlx"
)

// Tx wrapper struct.
type Tx struct {
	*sqlx.Tx
}

// Transaction wrapper.
// Transaction executes a method in a transaction.
func (t *Tx) Transaction(fn func(tx *Tx) (interface{}, error)) error {
	_, err := fn(t)
	return err
}
