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

package mongo

import (
	"reflect"

	"gopkg.in/mgo.v2"
)

// QueryBuilder builds complex queries
// managed by the document manager
type queryBuilder interface {
	Find(query interface{}) queryBuilder

	// Sort asks the database to order returned documents according to the
	// provided field names
	// @see http://www.mongodb.org/display/DOCS/Sorting+and+Natural+Order
	Sort(...string) queryBuilder

	// Limit restricts the maximum number of documents retrieved to n
	Limit(int) queryBuilder

	// Skip skips over the n initial documents from the query results.
	Skip(int) queryBuilder

	// Select enables selecting which fields should be retrieved for the results
	// found.
	// @see http://www.mongodb.org/display/DOCS/Retrieving+a+Subset+of+Fields
	Select(query interface{}) queryBuilder

	// One assigns one document or returns an error
	// it expects a struct pointer
	One(document interface{}) error

	// All assigns multiple documents in a slice or returns an error.
	// It expects a pointer to a slice of struct pointers
	All(documents interface{}) error
}

type defaultQueryBuilder struct {
	documentManager *defaultDocumentManager
	query           interface{}
	selection       interface{}
	limit, skip     int
	order           []string
}

func newDefaultQueryBuilder(documentManager *defaultDocumentManager) queryBuilder {
	return &defaultQueryBuilder{documentManager: documentManager}
}

func (qb *defaultQueryBuilder) Find(query interface{}) queryBuilder {
	qb.query = query
	return qb
}

func (qb *defaultQueryBuilder) Limit(limit int) queryBuilder {
	qb.limit = limit
	return qb
}

func (qb *defaultQueryBuilder) Skip(skip int) queryBuilder {
	qb.skip = skip
	return qb
}

func (qb *defaultQueryBuilder) Select(fieldSelection interface{}) queryBuilder {
	qb.selection = fieldSelection
	return qb
}

func (qb *defaultQueryBuilder) Sort(fields ...string) queryBuilder {
	qb.order = fields
	return qb
}

func (qb *defaultQueryBuilder) One(document interface{}) error {
	meta, err := qb.documentManager.metadatas.getMetadatas(reflect.TypeOf(document))
	if err != nil {
		return ErrDocumentNotRegistered
	}
	query := qb.buildQuery(meta.collectionName)
	if err := query.One(document); err != nil {
		return err
	}
	return qb.documentManager.resolveRelations(document)
}

func (qb *defaultQueryBuilder) All(documents interface{}) error {
	if value := reflect.ValueOf(documents); value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	} else if kind := value.Elem().Kind(); kind != reflect.Array && kind != reflect.Slice {
		return ErrNotAnArray
	}
	meta, err := qb.documentManager.metadatas.getMetadatas(reflect.TypeOf(documents).Elem().Elem())
	if err != nil {
		return ErrDocumentNotRegistered
	}
	query := qb.buildQuery(meta.collectionName)
	if err := query.All(documents); err != nil {
		return err
	}
	return qb.documentManager.resolveAllRelations(documents)
}

func (qb *defaultQueryBuilder) buildQuery(collectionName string) *mgo.Query {
	q := qb.documentManager.GetDB().C(collectionName).Find(qb.query)
	if qb.limit > 0 {
		q = q.Limit(qb.limit)
	}
	if qb.skip > 0 {
		q = q.Skip(qb.skip)
	}
	if len(qb.order) > 0 {
		q = q.Sort(qb.order...)
	}
	if qb.selection != nil {
		q = q.Select(qb.selection)
	}

	return q
}
