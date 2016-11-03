package db

import (
	"database/sql"

	"github.com/Mparaiso/go-tiger/logger"
)

type Transaction struct {
	*sql.Tx
	Logger logger.Logger
}

func (transaction *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	defer transaction.log(append([]interface{}{query}, args...)...)
	return transaction.Tx.Exec(query, args...)
}

func (transaction *Transaction) Rollback() (err error) {
	defer transaction.error("Rollback Transaction.", err)
	err = transaction.Tx.Rollback()
	return
}

func (transaction *Transaction) Commit() error {
	defer transaction.log("Commit Transaction.")
	return transaction.Tx.Commit()
}
func (transaction *Transaction) error(messages ...interface{}) {
	if transaction.Logger != nil {
		transaction.Logger.Log(logger.Error, messages...)
	}
}
func (transaction *Transaction) log(messages ...interface{}) {
	if transaction.Logger != nil {
		transaction.Logger.Log(logger.Debug, messages...)
	}
}
