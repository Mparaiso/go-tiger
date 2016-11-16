//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

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

func (transaction *Transaction) Prepare(query string) (*Statement, error) {
	defer transaction.log(query)
	stmt, err := transaction.Tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &Statement{statement: stmt, query: query, logger: transaction.Logger}, nil
}

func (transaction *Transaction) Query(query string, arguments ...interface{}) *Rows {
	defer transaction.log(query, arguments)
	return NewRows(transaction.Tx.Query(query, arguments...))
}

func (transaction *Transaction) QueryRow(query string, arguments ...interface{}) *Row {
	defer transaction.log(query, arguments)
	return NewRow(transaction.Tx.Query(query, arguments...))
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
