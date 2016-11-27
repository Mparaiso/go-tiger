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

// Package mongo provides an object document mapper, or ODM for mongodb, strongly influenced by Doctrine Mongo ODM.
package mongo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Mparaiso/go-tiger/funcs"
	"github.com/Mparaiso/go-tiger/logger"
	"github.com/Mparaiso/go-tiger/tag"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrDocumentNotRegistered is yielded when a type that has not been registered is requested by the DocumentManager
	ErrDocumentNotRegistered = fmt.Errorf("Error the type of the document was not registered in the document manager")
	// ErrIDFieldNotFound is yielded when _id field wasn't found in a struct
	ErrIDFieldNotFound = fmt.Errorf("Error no _id field defined for type")
	// ErrMappedFieldNotFound is yielded when the field of a mappedBy annotation was not found
	ErrMappedFieldNotFound = fmt.Errorf("Error mapped field not found, check mappedBy annotation for document")
	// ErrNotAstruct is yielded when a struct was expected
	ErrNotAstruct = fmt.Errorf("Error a struct was expected")
	// ErrNotAPointer is yielded when a pointer was expected
	ErrNotAPointer = fmt.Errorf("Error a pointer was expected")
	// ErrNotAnArray is yielded when an array was expected
	ErrNotAnArray = fmt.Errorf("Error an array was expected")
	// ErrNotImpletemented is yielded when a method was called yet is not implemented
	ErrNotImpletemented = fmt.Errorf("Error a called method is not implemented")
	// ErrFieldNotFound : Error a field metada was requested and not found
	ErrFieldNotFound = fmt.Errorf("Error a field metada was requested and not found ")
	zeroMetadata     = metadata{}
	zeroRelation     = relation{}
	// ZeroObjectID represents a zero value for bson.ObjectId
	zeroObjectID = reflect.Zero(reflect.TypeOf(bson.NewObjectId())).Interface().(bson.ObjectId)
)

// DocumentManager is a mongodb document manager
type DocumentManager interface {

	// Register a new document type, targetDocument is the name of the document and the collection name,
	// document is a pointer to struct.
	// returns an error on error.
	// use DocumentManager.RegisterMany to register many documents at the same time.
	Register(collectionName string, value interface{}) error

	// register many documents or returns an error on error
	RegisterMany(documents map[string]interface{}) error

	// Persist saves a document. No document is sent to the db
	// until flush is called
	Persist(document interface{})

	// Remove deletes a document. Flush must be called to commit changes
	// to the database
	Remove(document interface{})

	// Flush executes saves,updates and removes pending in the document manager
	Flush() error

	// FindID finds a document by ID
	FindID(id interface{}, returnValue interface{}) error

	// FIndOne finds a single document
	FindOne(query interface{}, returnValue interface{}) error

	// FindBy find documents by query
	FindBy(query interface{}, returnValues interface{}) error

	// FIndAll find all documents in a collection
	FindAll(returnValues interface{}) error

	// GetDB returns the driver's DB
	GetDB() *mgo.Database

	// SetLogger sets the logger
	SetLogger(logger.Logger)

	// CreateQuery creates a query builder for complex queries
	CreateQuery() QueryBuilder
}

type defaultDocumentManager struct {
	database  *mgo.Database
	metadatas metadatas
	tasks     tasks
	logger    logger.Logger
}

// NewDocumentManager returns a DocumentManager
func NewDocumentManager(database *mgo.Database) DocumentManager {
	return &defaultDocumentManager{database: database, metadatas: map[reflect.Type]metadata{}, tasks: tasks{}}
}

// GetDB returns the original mongodb connection
func (manager *defaultDocumentManager) GetDB() *mgo.Database {
	return manager.database
}

func (manager *defaultDocumentManager) SetLogger(Logger logger.Logger) {
	manager.logger = Logger
}

func (manager *defaultDocumentManager) log(messages ...interface{}) {
	if manager.logger != nil {
		manager.logger.Log(logger.Debug, messages...)
	}
}

// Register a new document type, targetDocument is the name of the document and the collection name,
// document is a pointer to struct.
// use DocumentManager.RegisterMany to register many documents at the same time.
func (manager *defaultDocumentManager) Register(targetDocument string, document interface{}) error {
	documentType := reflect.TypeOf(document)
	if documentType.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if documentType.Elem().Kind() != reflect.Struct {
		return ErrNotAstruct
	}
	meta, err := getTypeMetadatas(document)
	if err != nil {
		return err
	}
	meta.collectionName = targetDocument
	// parser := tag.NewParser(strings.NewReader(s string) )
	manager.metadatas[documentType] = meta

	manager.log("Type registered :", targetDocument, meta)
	return nil
}

func (manager *defaultDocumentManager) RegisterMany(documents map[string]interface{}) error {
	for targetDocument, document := range documents {
		if err := manager.Register(targetDocument, document); err != nil {
			return err
		}
	}
	return nil
}

func (manager *defaultDocumentManager) Persist(value interface{}) {
	if id, _ := manager.metadatas.getDocumentID(value); !id.Valid() {
		// new document, insert
		manager.metadatas.setIDForValue(value, bson.NewObjectId())
		manager.tasks[value] = insert
		return
	}
	// has an id, upsert
	manager.tasks[value] = update
}

func (manager *defaultDocumentManager) Remove(document interface{}) {
	manager.tasks[document] = del
}

func (manager *defaultDocumentManager) Flush() error {
	// TODO : a document should be flushed only once
	// keep track of a document that has already been flushed
	// and don't had it again to the tasks.
	// removing should take priority on persisting.
	for len(manager.tasks) != 0 {
		document, theTask := manager.tasks.pop()
		switch theTask {
		case del:
			if err := manager.doRemove(document); err != nil {
				return err
			}
		case insert, update:
			if err := manager.doPersist(document); err != nil {
				return err
			}
		}
	}
	return nil
}

func (manager *defaultDocumentManager) FindBy(query interface{}, documents interface{}) error {
	Value := reflect.ValueOf(documents)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if Value.Elem().Kind() != reflect.Array && Value.Elem().Kind() != reflect.Slice {
		return ErrNotAnArray
	}
	Type := Value.Elem().Type().Elem()
	meta, ok := manager.metadatas[Type]
	if !ok {
		return ErrDocumentNotRegistered
	}
	if err := manager.database.C(meta.collectionName).Find(query).All(documents); err != nil {
		return err
	}
	return manager.resolveAllRelations(documents)
}

func (manager *defaultDocumentManager) FindAll(documents interface{}) error {
	Value := reflect.ValueOf(documents)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if Value.Elem().Kind() != reflect.Array && Value.Elem().Kind() != reflect.Slice {
		return ErrNotAnArray
	}
	Type := Value.Elem().Type().Elem()
	meta, ok := manager.metadatas[Type]
	if !ok {
		return ErrDocumentNotRegistered
	}
	if err := manager.database.C(meta.collectionName).Find(nil).All(documents); err != nil {
		return err
	}
	return manager.resolveAllRelations(documents)
}

func (manager *defaultDocumentManager) FindOne(query interface{}, document interface{}) error {
	Value := reflect.ValueOf(document)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if Value.Elem().Kind() != reflect.Struct {
		return ErrNotAstruct
	}
	meta, ok := manager.metadatas[reflect.TypeOf(document)]
	if !ok {
		return ErrDocumentNotRegistered
	}
	if err := manager.database.C(meta.collectionName).Find(query).One(document); err != nil {
		return err
	}
	return manager.resolveRelations(document)
}

func (manager *defaultDocumentManager) FindID(documentID interface{}, document interface{}) error {
	Value := reflect.ValueOf(document)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if Value.Elem().Kind() != reflect.Struct {
		return ErrNotAstruct
	}
	meta, ok := manager.metadatas[reflect.TypeOf(document)]
	if !ok {
		return ErrDocumentNotRegistered
	}
	err := manager.database.C(meta.collectionName).FindId(documentID).One(document)
	if err != nil {
		return err
	}
	return manager.resolveRelations(document)
}

func (manager *defaultDocumentManager) CreateQuery() QueryBuilder {
	return newDefaultQueryBuilder(manager)
}

func (manager *defaultDocumentManager) structToMap(value interface{}) map[string]interface{} {
	// structToMap turns a struct into a map
	// ignored fields  and relations are ignored along with zero values if omitempty is configured
	result := map[string]interface{}{}
	Value := reflect.ValueOf(value)
	meta := manager.metadatas[Value.Type()]
	for _, field := range meta.fields {
		if field.ignore || (field.omitempty && isZero(Value.Elem().FieldByName(field.name).Interface())) {
			continue
		}
		if field.hasRelation() {
			continue
		}
		if field.name == meta.idField {
			result["_id"] = Value.Elem().FieldByName(field.name).Interface()
			continue
		}
		result[field.key] = Value.Elem().FieldByName(field.name).Interface()
	}
	return result
}

func (manager *defaultDocumentManager) doRemove(document interface{}) error {
	metadata, ok := manager.metadatas[reflect.TypeOf(document)]
	if !ok {
		return ErrDocumentNotRegistered
	}
	Value := reflect.Indirect(reflect.ValueOf(document))
	Map := manager.structToMap(document)
	if metadata.hasRelation() {
		for _, field := range metadata.getFieldsWithRelation() {
			if field.relation.cascade == all || field.relation.cascade == remove {
				switch field.relation.relation {
				case referenceMany:
					meta, Type := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if Type != nil {
						many := Value.FieldByName(field.name)
						for i := 0; i < many.Len(); i++ {
							doc := many.Index(i)
							idField, ok := meta.findIDField()
							if !ok {
								continue
							}
							id := doc.Elem().FieldByName(idField.name)
							if !isZero(id.Interface()) {
								manager.tasks[doc.Interface()] = del
							}
						}
					}
				case referenceOne:
					// add id of the reference to map , and add the reference in the documents to be saved
					meta, Type := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if Type != nil {
						one := Value.FieldByName(field.name)
						if isZero(one.Interface()) {
							continue
						}
						idField, ok := meta.findIDField()
						if !ok {
							continue
						}
						id := one.Elem().FieldByName(idField.name)
						if !isZero(id.Interface()) {
							manager.tasks[one.Interface()] = del
						}
					}
				}
			}
		}
	}
	err := manager.database.C(metadata.collectionName).RemoveId(Map["_id"])
	if err != nil {
		return err
	}
	// set the id to a zero value
	manager.metadatas.setIDForValue(document, zeroObjectID)
	manager.log(fmt.Sprintf("Removed document with id '%s' from collection '%s' ", Map["_id"], metadata.collectionName))
	return nil
}

func (manager *defaultDocumentManager) doPersist(document interface{}) error {
	metadata, ok := manager.metadatas[reflect.TypeOf(document)]
	if !ok {
		return ErrDocumentNotRegistered
	}
	Value := reflect.Indirect(reflect.ValueOf(document))
	Map := manager.structToMap(document)
	if metadata.hasRelation() {
		for _, field := range metadata.getFieldsWithRelation() {
			if field.relation.relationMap != mappedBy {
				switch field.relation.relation {
				case referenceMany:
					objectIDs := []bson.ObjectId{}
					meta, Type := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if Type != nil {
						many := Value.FieldByName(field.name)
						for i := 0; i < many.Len(); i++ {
							doc := many.Index(i)
							idField, ok := meta.findIDField()
							if !ok {
								continue
							}
							id := doc.Elem().FieldByName(idField.name)
							if isZero(id.Interface()) {
								doc.Elem().FieldByName(idField.name).Set(reflect.ValueOf(bson.NewObjectId()))
							}
							objectIDs = append(objectIDs, doc.Elem().FieldByName(idField.name).Interface().(bson.ObjectId))
							if field.relation.cascade == all || field.relation.cascade == persist {
								manager.tasks[doc.Interface()] = insert
							}
						}
					}
					Map[field.key] = objectIDs
				case referenceOne:
					// add id of the reference to map , and add the reference in the documents to be saved
					meta, Type := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if Type != nil {
						one := Value.FieldByName(field.name)
						if isZero(one.Interface()) {
							continue
						}
						idField, ok := meta.findIDField()
						if !ok {
							continue
						}
						id := one.Elem().FieldByName(idField.name)
						if isZero(id.Interface()) {
							one.Elem().FieldByName(idField.name).Set(reflect.ValueOf(bson.NewObjectId()))
						}
						if field.relation.cascade == all || field.relation.cascade == persist {
							manager.tasks[one.Interface()] = insert
						}
						Map[field.key] = one.Elem().FieldByName(idField.name).Interface().(bson.ObjectId)
					}
				}
			}
		}
	}
	id := Map["_id"]
	if changeInfo, err := manager.database.C(metadata.collectionName).UpsertId(id, bson.M{"$set": stripID(Map)}); err != nil {
		return err
	} else {
		manager.log(fmt.Sprintf("Persisted document with id '%s' from collection '%s' , %+v ", id, metadata.collectionName, changeInfo))
	}
	return nil
}

func (manager *defaultDocumentManager) resolveAllRelations(documents interface{}) error {
	// this operation is recursive so we need to keep track of the documents than have already
	// been fetched from the DB by their (unique) objectIDs.
	// the relations are resolved recursively. When no relation needs to be resolved or if an error occurs, return.
	return manager.doResolveAllRelations(documents, map[bson.ObjectId]interface{}{})
}

func (manager *defaultDocumentManager) resolveRelations(document interface{}) error {
	return manager.doResolveRelations(document, map[bson.ObjectId]interface{}{})
}

func (manager *defaultDocumentManager) doResolveAllRelations(documents interface{}, fetchedDocuments map[bson.ObjectId]interface{}) error {
	manager.log("Resolving all relations for :", reflect.TypeOf(documents))
	Pointer := reflect.ValueOf(documents)
	// expect a pointer
	if Pointer.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Collection := Pointer.Elem()
	// expect an array or a slice
	if kind := Collection.Kind(); kind != reflect.Array && kind != reflect.Slice {
		return ErrNotAnArray
	}
	// find metadata for collection element type
	meta, err := manager.metadatas.getMetadatas(Collection.Type().Elem())
	if err != nil {
		return err
	}
	// get an []reflect.Value so it is easy to iterate on reflect.Value
	values := func() []reflect.Value {
		res := []reflect.Value{}
		for i := 0; i < Collection.Len(); i++ {
			res = append(res, Collection.Index(i))
		}
		return res
	}()
	// key values by bson.Object so they are easier to look up
	valuesKeyedByObjectID := keyValuesByObjectID(values, func(val reflect.Value) bson.ObjectId {
		id, _ := manager.metadatas.getDocumentID(val.Interface())
		return id
	})
	// add values to previously fetched objects
	for objectID, value := range valuesKeyedByObjectID {
		fetchedDocuments[objectID] = value.Interface()
	}
	// if the metadata has relations
	if meta.hasRelation() {

		// get all objects ids
		documentIds := getObjectIds(valuesKeyedByObjectID)
		// for each field that has a relation
		for _, field := range meta.getFieldsWithRelation() {
			manager.log("\tRelation for field : ", field.name, field.relation.relation, field.relation.targetDocument, field.relation.relationMap, field.relation.relationMapField)
			switch field.relation.relation {

			case referenceMany:
				// the documents reference many related documents
				results := []map[string]interface{}{}
				if err = manager.GetDB().C(meta.collectionName).Find(bson.M{"_id": bson.M{"$in": documentIds}}).Select(bson.M{field.key: 1, "_id": 1}).All(&results); err != nil {
					return err
				}
				resultsKeyedByObjectID := keyResultsBySourceID(results, func(result map[string]interface{}) bson.ObjectId {
					return result["_id"].(bson.ObjectId)
				})

				// let's see if some related documents have already been fetched
				for objectID, result := range resultsKeyedByObjectID {
					value := valuesKeyedByObjectID[objectID]
					// if field is empty continue
					if _, ok := result[field.key]; !ok {
						continue
					}
					// otherwise iterate
					for _, relatedID := range result[field.key].([]interface{}) {
						if document, ok := fetchedDocuments[relatedID.(bson.ObjectId)]; ok {
							value.Elem().FieldByName(field.name).Set(reflect.Append(value.Elem().FieldByName(field.name), reflect.ValueOf(document)))
						}
					}
				}
				// let's filter out already existing related documents by objectID
				relatedObjectIds :=
					filterObjectIds(
						mapInterfacesToObjectIds(
							flattenSliceOfInterfaces(
								mapResultsToInterfaces(results, func(result map[string]interface{}) []interface{} {
									if _, ok := result[field.key]; !ok {
										return []interface{}{}
									}
									return result[field.key].([]interface{})
								}),
							), func(i interface{}) bson.ObjectId {
								return i.(bson.ObjectId)
							}),
						func(id bson.ObjectId) bool {
							_, ok := fetchedDocuments[id]
							return !ok
						})
				// if there is no related document to fetch, continue
				if len(relatedObjectIds) == 0 {
					continue
				}
				_, relatedType := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
				if relatedType == nil {
					return ErrDocumentNotRegistered
				}
				relatedDocumentValues := reflect.New(reflect.SliceOf(relatedType))
				// fetch the remaining related documents
				if err = manager.GetDB().C(field.relation.targetDocument).Find(bson.M{"_id": bson.M{"$in": relatedObjectIds}}).All(relatedDocumentValues.Interface()); err != nil {
					return err
				}
				for objectID, result := range resultsKeyedByObjectID {
					value := valuesKeyedByObjectID[objectID]
					for _, id := range result[field.key].([]interface{}) {
						for i := 0; i < relatedDocumentValues.Elem().Len(); i++ {
							relatedObjectID, _ := manager.metadatas.getDocumentID(relatedDocumentValues.Elem().Index(i).Interface())
							if id.(bson.ObjectId) == relatedObjectID {
								value.Elem().FieldByName(field.name).Set(reflect.Append(value.Elem().FieldByName(field.name), relatedDocumentValues.Elem().Index(i)))
							}
						}
					}
				}
				// lets resolve the relations of the related documents
				if err = manager.doResolveAllRelations(relatedDocumentValues.Interface(), fetchedDocuments); err != nil {
					return err
				}
			case referenceOne:
				switch field.relation.relationMap {

				case mappedBy:

					// first we need to search the owning side for metadata , the owning side is defined by the argument of mappedBy
					relatedMeta, relatedType := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if relatedType == nil {
						return ErrDocumentNotRegistered
					}
					relatedDocumentMaps := []map[string]interface{}{}
					// We need the related struct field and the mongodb key of the owning side which holds the reference to the source document
					relatedField, found := relatedMeta.findFieldByFieldName(field.relation.relationMapField)
					if !found {
						return ErrFieldNotFound
					}
					// we have a list of source document ids, let's fetch the related documents
					if err = manager.GetDB().C(relatedMeta.collectionName).Find(bson.M{"_id": bson.M{"$nin": documentIds}, relatedField.key: bson.M{"$in": documentIds}}).Select(bson.M{"_id": 1, relatedField.key: 1}).All(&relatedDocumentMaps); err != nil && err != mgo.ErrNotFound {
						return err
					}
					// 2 cases here. if the related documents reference many then we need to search through an array
					// if the related documents reference one ,then it is a single value
					relatedDocumentsMapsMappedByDocumentID := map[bson.ObjectId]map[string]interface{}{}
					relatedDocumentIds := []bson.ObjectId{}
					switch relatedField.relation.relation {
					case referenceMany:
						for _, relatedDocument := range relatedDocumentMaps {
							// only append to relatedDocumentIds the documents that have not been fetched yet
							if _, ok := fetchedDocuments[relatedDocument["_id"].(bson.ObjectId)]; !ok {
								relatedDocumentIds = append(relatedDocumentIds, relatedDocument["_id"].(bson.ObjectId))
							}
							for _, id := range relatedDocument[relatedField.key].([]interface{}) {
								relatedDocumentsMapsMappedByDocumentID[id.(bson.ObjectId)] = relatedDocument
							}
						}
					default:
						for _, relatedDocument := range relatedDocumentMaps {
							// only append to relatedDocumentIds the documents that have not been fetched yet
							if _, ok := fetchedDocuments[relatedDocument["_id"].(bson.ObjectId)]; !ok {
								relatedDocumentIds = append(relatedDocumentIds, relatedDocument["_id"].(bson.ObjectId))
							}
							relatedDocumentsMapsMappedByDocumentID[relatedDocument[relatedField.key].(bson.ObjectId)] = relatedDocument
						}
					}

					// let's load the actual related documents fully typed
					relatedDocuments := reflect.New(reflect.SliceOf(relatedType))
					if err = manager.GetDB().C(relatedMeta.collectionName).Find(bson.M{"_id": bson.M{"$in": relatedDocumentIds}}).All(relatedDocuments.Interface()); err != nil && err != mgo.ErrNotFound {
						return err
					}
					relatedDocumentsMappedByDocumentID := map[bson.ObjectId]reflect.Value{}
					// let's first add the documents that have already been fetched
					for documentId, relatedDocumentMap := range relatedDocumentsMapsMappedByDocumentID {
						if document, ok := fetchedDocuments[relatedDocumentMap["_id"].(bson.ObjectId)]; ok {
							relatedDocumentsMappedByDocumentID[documentId] = reflect.ValueOf(document)
						}
					}
					// let's now add the new related documents we just fetched
					for i := 0; i < relatedDocuments.Elem().Len(); i++ {
						for documentId, relatedDocumentMap := range relatedDocumentsMapsMappedByDocumentID {
							if relatedDocumentMap["_id"].(bson.ObjectId) == relatedDocuments.Elem().Index(i).Elem().FieldByName(relatedMeta.idField).Interface().(bson.ObjectId) {
								relatedDocumentsMappedByDocumentID[documentId] = relatedDocuments.Elem().Index(i)
							}
						}
					}
					// let's now add each related mapped document
					for documentID, value := range valuesKeyedByObjectID {
						value.Elem().FieldByName(field.name).Set(relatedDocumentsMappedByDocumentID[documentID])
					}
					// let's resolve the possible relations in the related documents we just fetched
					if err = manager.doResolveAllRelations(relatedDocuments.Interface(), fetchedDocuments); err != nil {
						return err
					}

				default:

					// the documents reference one related document
					results := []map[string]interface{}{}
					if err = manager.GetDB().C(meta.collectionName).Find(bson.M{"_id": bson.M{"$in": documentIds}}).Select(bson.M{field.key: 1, "_id": 1}).All(&results); err != nil && err != mgo.ErrNotFound {
						return err
					}
					resultsKeyedByObjectID := keyResultsBySourceID(results, func(result map[string]interface{}) bson.ObjectId {
						return result["_id"].(bson.ObjectId)
					})

					// search in fetched documents if the relation can already be satisified
					// if yes then set the field of the related doc to the fetched document
					for objectID, result := range resultsKeyedByObjectID {
						relatedObjectID := result[field.key].(bson.ObjectId)
						if document, ok := fetchedDocuments[relatedObjectID]; ok {
							valuesKeyedByObjectID[objectID].Elem().FieldByName(field.name).Set(reflect.ValueOf(document))
						}
					}
					// we don't need the object ids that have already been fetched
					relatedObjectIds := filterObjectIds(mapResultsToRelatedObjectIds(results, func(result map[string]interface{}) bson.ObjectId {
						return result[field.key].(bson.ObjectId)
					}), func(id bson.ObjectId) bool {
						_, ok := fetchedDocuments[id]
						return !ok
					})
					// if there is no related document left to fetch , continue
					if len(relatedObjectIds) == 0 {
						continue
					}
					relatedMeta, relatedType := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if relatedType == nil {
						return ErrDocumentNotRegistered
					}
					relatedDocumentValues := reflect.New(reflect.SliceOf(relatedType))
					// fetch the remaining documents from the db
					if err = manager.GetDB().C(field.relation.targetDocument).Find(bson.M{"_id": bson.M{"$in": relatedObjectIds}}).All(relatedDocumentValues.Interface()); err != nil && err != mgo.ErrNotFound {
						return err
					}
					relatedDocumentValuesKeyedByObjectID := keyRelatedResultsByObjectID(func() []reflect.Value {
						// transform reflect.Value into []reflect.Value so it can be iterated more easily
						values := []reflect.Value{}
						for i := 0; i < relatedDocumentValues.Elem().Len(); i++ {
							values = append(values, relatedDocumentValues.Elem().Index(i))
						}
						return values
					}(), func(value reflect.Value) bson.ObjectId {
						//println(relatedMeta.idField)
						//println(value.Elem().FieldByName(relatedMeta.idField).Interface().(bson.ObjectId))
						return value.Elem().FieldByName(relatedMeta.idField).Interface().(bson.ObjectId)
					})
					for id, value := range valuesKeyedByObjectID {
						result := resultsKeyedByObjectID[id]
						relatedID := result[field.key].(bson.ObjectId)
						relatedResult := relatedDocumentValuesKeyedByObjectID[relatedID]
						value.Elem().FieldByName(field.name).Set(relatedResult)
					}
					// lets resolve the relations of the related documents
					if err = manager.doResolveAllRelations(relatedDocumentValues.Interface(), fetchedDocuments); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (manager *defaultDocumentManager) doResolveRelations(document interface{}, fetchedDocuments map[bson.ObjectId]interface{}) error {
	manager.log("Resolving relation for :", reflect.TypeOf(document))
	Value := reflect.ValueOf(document)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	meta, err := manager.metadatas.getMetadatas(reflect.TypeOf(document))
	if err != nil {
		return ErrDocumentNotRegistered
	}
	documentID, err := manager.metadatas.getDocumentID(document)
	if err != nil {
		return err
	}
	if _, ok := fetchedDocuments[documentID]; ok {
		// document has already been fetched ! just return
		return nil
	} else {
		// add to the already fetched document list
		fetchedDocuments[documentID] = document
	}
	if meta.hasRelation() {
		for _, field := range meta.getFieldsWithRelation() {
			manager.log("\tRelation for field : ", field.name, field.relation.relation, field.relation.targetDocument, field.relation.relationMap, field.relation.relationMapField)
			switch field.relation.relation {
			case referenceMany:
				ids := []interface{}{}
				relatedMeta, Type := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
				if Type == nil {
					return ErrDocumentNotRegistered
				}
				switch field.relation.relationMap {
				case mappedBy:
					relatedDocuments := []map[string]interface{}{}
					mappedCollection := field.relation.targetDocument
					mappedField := field.relation.relationMapField
					fieldmetadata, found := relatedMeta.findFieldByFieldName(mappedField)
					if !found {
						return ErrMappedFieldNotFound
					}
					if err := manager.GetDB().C(mappedCollection).Find(bson.M{fieldmetadata.key: documentID}).Select(bson.M{"_id": 1}).All(&relatedDocuments); err != nil && err != mgo.ErrNotFound {
						return err
					} else if err == mgo.ErrNotFound {
						continue
					}
					for _, relatedDocument := range relatedDocuments {
						ids = append(ids, relatedDocument["_id"])
					}
				default:
					fetchedIds := map[string]interface{}{}
					if err = manager.GetDB().C(meta.collectionName).FindId(documentID).Select(bson.M{field.key: 1}).One(&fetchedIds); err != nil {
						return err
					}
					if _, ok := fetchedIds[field.key]; !ok {
						continue
					}
					ids = fetchedIds[field.key].([]interface{})
				}
				// if no related document, continue
				if len(ids) == 0 {
					continue
				}

				relatedResults := reflect.New(reflect.SliceOf(Type))
				if err = manager.GetDB().C(relatedMeta.collectionName).Find(bson.M{"_id": bson.M{"$in": ids}}).All(relatedResults.Interface()); err != mgo.ErrNotFound && err != nil {
					return err
				}
				Value.Elem().FieldByName(field.name).Set(relatedResults.Elem())
				if err = manager.doResolveAllRelations(relatedResults.Interface(), fetchedDocuments); err != nil {
					return err
				}
			case referenceOne:
				switch field.relation.relationMap {
				case mappedBy:
					relatedMeta, Type := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if Type == nil {
						return ErrDocumentNotRegistered
					}
					// let's try to find the related document in the fetchedDocuments
					found := false
					for _, document := range fetchedDocuments {
						documentValue := reflect.ValueOf(document)
						if documentValue.Type() == Type {
							if documentValue.Elem().FieldByName(field.relation.relationMapField).Interface().(bson.ObjectId) == documentID {
								Value.Elem().FieldByName(field.name).Set(documentValue)
								found = true
								break
							}
						}
					}
					// we found the related document, continue with the next relation
					if found == true {
						continue
					}
					// we need the related document's key which holds the id of document
					relatedField, found := relatedMeta.findFieldByFieldName(field.relation.relationMapField)
					if found == false {
						return ErrFieldNotFound
					}
					relatedDocument := reflect.New(Type.Elem())
					// let's search for the related document by the source documentId
					if err = manager.GetDB().C(relatedMeta.collectionName).Find(bson.M{relatedField.key: documentID}).One(relatedDocument.Interface()); err != mgo.ErrNotFound && err != nil {
						return err
					} else if err != mgo.ErrNotFound {
						Value.Elem().FieldByName(field.name).Set(relatedDocument)
						if err = manager.doResolveRelations(relatedDocument.Interface(), fetchedDocuments); err != nil {
							return err
						}
					}
				default:
					// we need to requery the main document to get the id of the related document
					documentMap := map[string]interface{}{}
					err = manager.GetDB().C(meta.collectionName).FindId(documentID).Select(bson.M{field.key: 1}).One(&documentMap)
					if err != nil {
						return err
					}
					if documentMap[field.key] == nil {
						// no referenced document was found
						continue
					}
					id := documentMap[field.key].(bson.ObjectId)

					relatedMeta, Type := manager.metadatas.findMetadataByCollectionName(field.relation.targetDocument)
					if Type == nil {
						return ErrDocumentNotRegistered
					}
					relatedDocument := reflect.New(Type.Elem())
					if err = manager.GetDB().C(relatedMeta.collectionName).FindId(id).One(relatedDocument.Interface()); err != mgo.ErrNotFound && err != nil {
						return err
					} else if err != mgo.ErrNotFound {
						Value.Elem().FieldByName(field.name).Set(relatedDocument)
						if err = manager.doResolveRelations(relatedDocument.Interface(), fetchedDocuments); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

type metadata struct {
	collectionName string
	idField        string
	fields         []field
}

func (meta metadata) String() string {
	return metadataToString(meta)

}
func (meta metadata) findFieldByFieldName(fieldname string) (f field, found bool) {
	for _, field := range meta.fields {
		if field.name == fieldname {
			return field, true
		}
	}
	return f, false
}
func (meta metadata) findIDField() (f field, found bool) {
	for _, field := range meta.fields {
		if field.key == "_id" {
			return field, true
		}
	}
	return
}

// hasRelation returns true if one of the fields has a relation
func (meta metadata) hasRelation() bool {
	for _, field := range meta.fields {
		if field.hasRelation() {
			return true
		}
	}
	return false
}

// getFieldsWithRelation returns a collection of fields with relations
func (meta metadata) getFieldsWithRelation() (fields []field) {
	for _, field := range meta.fields {
		if field.hasRelation() {
			fields = append(fields, field)
		}
	}
	return
}

type metadatas map[reflect.Type]metadata

// GetMetadatas returns the metada for a given type
func (metas metadatas) getMetadatas(Type reflect.Type) (metadata, error) {
	if meta, ok := metas[Type]; !ok {
		return zeroMetadata, ErrDocumentNotRegistered
	} else {
		return meta, nil
	}

}

func (metas metadatas) String() string {
	return fmt.Sprintf("%+v", metas)
}
func (metas metadatas) setIDForValue(document interface{}, id bson.ObjectId) error {
	Value := reflect.ValueOf(document)
	meta, ok := metas[Value.Type()]
	if !ok {
		return ErrDocumentNotRegistered
	}
	Value.Elem().FieldByName(meta.idField).Set(reflect.ValueOf(id))
	return nil
}

// GetIDForValue returns the value of the id field for document
func (metas metadatas) getDocumentID(document interface{}) (id bson.ObjectId, err error) {
	Value := reflect.ValueOf(document)
	meta, ok := metas[Value.Type()]
	if !ok {
		return id, ErrDocumentNotRegistered
	}
	idFields, ok := meta.findIDField()
	if !ok {
		return id, ErrIDFieldNotFound
	}
	return Value.Elem().FieldByName(idFields.name).Interface().(bson.ObjectId), nil

}
func (metas metadatas) findMetadataByCollectionName(name string) (metadata, reflect.Type) {
	for Type, meta := range metas {
		if meta.collectionName == name {
			return meta, Type
		}
	}
	return zeroMetadata, nil
}

type field struct {
	// mongodb document key
	key string
	// struct field name
	name string
	// whether to omit 0 values
	omitempty bool
	relation  relation
	ignore    bool
}

func (f field) String() string {
	return fmt.Sprintf(" key:'%s', name:'%s', omitempty:'%v' ignore:'%v' relation:%s ",
		f.key, f.name, f.omitempty, f.ignore, f.relation)

}

func (f field) hasRelation() bool {
	return f.relation != zeroRelation
}

type relation struct {
	relation         relationType
	targetDocument   string
	cascade          cascade
	relationMap      relationMap
	relationMapField string
}

func (r relation) String() string {
	if isZero(r) {
		return "{}"
	}
	return fmt.Sprintf("{ relation: '%s', targetDocument: '%s', cascade: '%v', map: '%s', mapField: '%v' } ",
		r.relation, r.targetDocument, r.cascade, r.relationMap, r.relationMapField)
}

type relationMap int

const (
	_ relationMap = iota
	mappedBy
	inversedBy
)

func (m relationMap) String() string {
	switch m {
	case mappedBy:
		return "mappedBy"
	case inversedBy:
		return "inversedBy"
	default:
		return ""
	}
}

type relationType int

const (
	_ relationType = iota
	referenceMany
	referenceOne
)

func (Type relationType) String() string {
	switch Type {
	case referenceMany:
		return "referenceMany"
	case referenceOne:
		return "referenceOne"
	}
	return ""
}

type cascade int

const (
	_ cascade = iota
	all
	persist
	remove
)

type task int

const (
	del task = iota
	insert
	update
)

type tasks map[interface{}]task

func (t tasks) pop() (interface{}, task) {
	for value, task := range t {
		delete(t, value)
		return value, task
	}
	return nil, 0
}
func stripID(Map map[string]interface{}) map[string]interface{} {
	delete(Map, "_id")
	return Map
}

func isZero(value interface{}) bool {
	Value := reflect.ValueOf(value)
	if Value.Kind() == reflect.Array || Value.Kind() == reflect.Slice {
		if Value.Len() == 0 {
			return true
		}
		return false
	}
	return Value.Interface() == reflect.Zero(Value.Type()).Interface()
}

func isPointer(value interface{}) bool {
	return reflect.ValueOf(value).Kind() == reflect.Ptr
}

func isStruct(value interface{}) bool {
	return reflect.ValueOf(value).Kind() == reflect.Struct
}

func isIterable(value interface{}) bool {
	kind := reflect.ValueOf(value).Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

func getTypeMetadatas(value interface{}) (meta metadata, err error) {
	Value := reflect.Indirect(reflect.ValueOf(value))
	Type := Value.Type()
	// iterate through struct fields
	for i := 0; i < Value.NumField(); i++ {
		Field := Type.Field(i)
		MetaField := field{name: Field.Name, key: strings.ToLower(Field.Name)}
		Tag := Field.Tag.Get("bson")
		parts := strings.Split(Tag, ",")
		if len(parts) > 0 {
			if key := strings.TrimSpace(parts[0]); key != "" {
				MetaField.key = key
				if key == "_id" {
					meta.idField = Field.Name
				}
			}
		}
		if len(parts) > 1 {
			if part := strings.TrimSpace(parts[1]); part == "omitempty" {
				MetaField.omitempty = true
			}
		}
		Tag = Field.Tag.Get("odm")
		if Tag == "-" {
			MetaField.ignore = true
			continue
		}
		parser := tag.NewParser(strings.NewReader(Tag))
		var definitions []*tag.Definition
		definitions, err = parser.Parse()
		if err != nil {
			return meta, err
		}
		for _, definition := range definitions {
			switch strings.ToLower(definition.Name) {

			case "id":
				meta.idField = Field.Name
			case "key":
				MetaField.key = definition.Value
			case "omitempty":
				MetaField.omitempty = true
			case "referencemany", "referenceone":

				Relation := relation{}
				switch strings.ToLower(definition.Name) {
				case "referencemany":
					Relation.relation = referenceMany
				case "referenceone":
					Relation.relation = referenceOne
				}
				for _, parameter := range definition.Parameters {
					switch strings.ToLower(parameter.Key) {
					case "mappedby":
						Relation.relationMap = mappedBy
						Relation.relationMapField = parameter.Value
					case "targetdocument":
						Relation.targetDocument = parameter.Value
					case "cascade":
						switch strings.ToLower(parameter.Value) {
						case "persist":
							Relation.cascade = persist
						case "remove":
							Relation.cascade = remove
						case "all":
							Relation.cascade = all
						}
					}
					MetaField.relation = Relation
				}
			}
		}
		meta.fields = append(meta.fields, MetaField)
	}
	return
}

func metadataToString(meta metadata) string {
	result := "metadata : {"
	result += "collectionName: '" + meta.collectionName + "', "
	result += "idField: '" + meta.idField + "' "
	result += "fields :[\n"
	for i, field := range meta.fields {
		if i > 0 {
			result += ",\n "
		}
		result += "{" + field.String() + "}"
	}
	return result + "\n]}\n"
}

var (
	keyValuesByObjectID func(collection []reflect.Value, selector func(reflect.Value) bson.ObjectId) map[bson.ObjectId]reflect.Value

	_ = funcs.Must(funcs.MakeKeyBy(&keyValuesByObjectID))

	getObjectIds func(map[bson.ObjectId]reflect.Value) []bson.ObjectId

	_ = funcs.Must(funcs.MakeGetKeys(&getObjectIds))

	flattenSliceOfInterfaces func([][]interface{}) []interface{}

	_ = funcs.Must(funcs.MakeFlatten(&flattenSliceOfInterfaces))

	mapResultsToInterfaces func([]map[string]interface{}, func(map[string]interface{}) []interface{}) [][]interface{}

	_ = funcs.Must(funcs.MakeMap(&mapResultsToInterfaces))

	keyResultsBySourceID func(results []map[string]interface{}, mapper func(result map[string]interface{}) (id bson.ObjectId)) map[bson.ObjectId]map[string]interface{}

	_ = funcs.Must(funcs.MakeKeyBy(&keyResultsBySourceID))

	mapResultsToRelatedObjectIds func(results []map[string]interface{}, mapper func(result map[string]interface{}) bson.ObjectId) []bson.ObjectId

	_ = funcs.Must(funcs.MakeMap(&mapResultsToRelatedObjectIds))

	keyRelatedResultsByObjectID func(results []reflect.Value, mapper func(result reflect.Value) bson.ObjectId) map[bson.ObjectId]reflect.Value

	_ = funcs.Must(funcs.MakeKeyBy(&keyRelatedResultsByObjectID))

	mapInterfacesToObjectIds func([]interface{}, func(interface{}) bson.ObjectId) []bson.ObjectId

	_ = funcs.Must(funcs.MakeMap(&mapInterfacesToObjectIds))

	filterObjectIds func([]bson.ObjectId, func(id bson.ObjectId) bool) []bson.ObjectId

	_ = funcs.Must(funcs.MakeFilter(&filterObjectIds))

	indexOf func([]interface{}, interface{}) int

	_ = funcs.Must(funcs.MakeIndexOf(&indexOf))
)
