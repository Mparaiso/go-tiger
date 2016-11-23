package mongo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Mparaiso/go-tiger/logger"
	"github.com/Mparaiso/go-tiger/tag"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrDocumentNotRegistered is yield when a type that has not been registered is requested by the DocumentManager
	ErrDocumentNotRegistered = fmt.Errorf("Error the type of the document was not registered in the document manager")
	// ErrIDFieldNotFound is yield when _id field wasn't found in a struct
	ErrIDFieldNotFound = fmt.Errorf("Error no _id field defined for type")
	// ErrNotAstruct is yield when a struct was expected
	ErrNotAstruct = fmt.Errorf("Error a struct was expected")
	// ErrNotAPointer is yield when a pointer was expected
	ErrNotAPointer = fmt.Errorf("Error a pointer was expected")
	// ErrNotAnArray is yield when an array was expected
	ErrNotAnArray = fmt.Errorf("Error an array was expected")
	// ErrNotImpletemented is yield when a method was called yet is not implemented
	ErrNotImpletemented = fmt.Errorf("Error a called method is not implemented")

	zeroMetadata = metadata{}
	zeroRelation = relation{}
)

type metadata struct {
	collectionName string
	idField        string
	fields         []field
}

func (meta metadata) String() string {
	return metadataToString(meta)

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

func (metas metadatas) String() string {
	return fmt.Sprintf("%+v", metas)
}
func (metas metadatas) SetIDForValue(document interface{}, id bson.ObjectId) error {
	Value := reflect.ValueOf(document)
	meta, ok := metas[Value.Type()]
	if !ok {
		return ErrDocumentNotRegistered
	}
	Value.Elem().FieldByName(meta.idField).Set(reflect.ValueOf(id))
	return nil
}

// GetIDForValue returns the value of the id field for document
func (metas metadatas) GetIDForValue(document interface{}) (id bson.ObjectId, err error) {
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
func (metas metadatas) FindMetadataByCollectionName(name string) (metadata, reflect.Type) {
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
	return fmt.Sprintf(" key:'%s', name:'%s', omitempty:'%v' ignore:'%v' relation:%s ", f.key, f.name, f.omitempty, f.ignore, f.relation)
}

func (f field) hasRelation() bool {
	return f.relation != zeroRelation
}

type relation struct {
	relation       relationType
	targetDocument string
	cascade        cascade
}

func (r relation) String() string {
	if isZero(r) {
		return "{}"
	}
	return fmt.Sprintf("{ relation: '%s', targetDocument: '%s', cascade: '%v' } ", r.relation, r.targetDocument, r.cascade)
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

// DocumentManager is a mongodb document manager
type DocumentManager interface {

	// Register adds a new type to the document manager.
	// A pointer to struct is expected
	Register(collectionName string, value interface{}) error

	// Persist saves a document. No document is sent to the db
	// until flush is called
	Persist(value interface{})

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
func (manager *defaultDocumentManager) Register(collectionName string, value interface{}) error {
	if !isPointer(value) {
		return ErrNotAPointer
	}
	meta, err := getTypeMetadatas(value)
	if err != nil {
		return err
	}
	meta.collectionName = collectionName
	// parser := tag.NewParser(strings.NewReader(s string) )
	manager.metadatas[reflect.TypeOf(value)] = meta

	manager.log("Type registered :", collectionName, "\r", meta)
	return nil
}

func (manager *defaultDocumentManager) Persist(value interface{}) {
	if _, err := manager.metadatas.GetIDForValue(value); err != nil {
		// new document, insert
		manager.metadatas.SetIDForValue(value, bson.NewObjectId())
		manager.tasks[value] = insert
		return
	}
	// has an id, upsert
	manager.tasks[value] = update
}

func (manager *defaultDocumentManager) structToMap(value interface{}) map[string]interface{} {
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

		// unfortunalty mgo/bson lowercase fields. in order for the fields to be fetched back
		// easily we need to lower case our fields too.
		result[field.key] = Value.Elem().FieldByName(field.name).Interface()
	}
	return result
}

func (manager *defaultDocumentManager) Flush() error {
	for len(manager.tasks) != 0 {
		document, theTask := manager.tasks.pop()
		metadata, ok := manager.metadatas[reflect.TypeOf(document)]
		if !ok {
			return ErrDocumentNotRegistered
		}
		Value := reflect.Indirect(reflect.ValueOf(document))
		Map := manager.structToMap(document)
		if metadata.hasRelation() {
			for _, field := range metadata.getFieldsWithRelation() {
				if field.relation.cascade == all ||
					(field.relation.cascade == persist && theTask == insert) ||
					(field.relation.cascade == persist && theTask == update) ||
					(field.relation.cascade == remove && theTask == del) {
					switch field.relation.relation {
					case referenceMany:
						objectIDs := []bson.ObjectId{}
						meta, Type := manager.metadatas.FindMetadataByCollectionName(field.relation.targetDocument)
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

								manager.tasks[doc.Interface()] = theTask
							}
						}
						Map[field.key] = objectIDs
					case referenceOne:
						// add id of the reference to map , and add the reference in the documents to be saved
						meta, Type := manager.metadatas.FindMetadataByCollectionName(field.relation.targetDocument)
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
							manager.tasks[one.Interface()] = theTask
							Map[field.key] = one.Elem().FieldByName(idField.name).Interface().(bson.ObjectId)
						}
					}
				}
			}
		}
		switch theTask {

		case insert:
			if err := manager.database.C(metadata.collectionName).Insert(Map); err != nil {
				return err
			}
		case update:
			if changeInfo, err := manager.database.C(metadata.collectionName).UpsertId(Map["_id"], bson.M{"$set": stripID(Map)}); err != nil {
				return err
			} else {
				manager.log("upsert", fmt.Sprintf("%+v", changeInfo))
			}
		case del:
			if err := manager.GetDB().C(metadata.collectionName).RemoveId(Map["_id"]); err != nil {
				return err
			}
		}

	}
	return nil
}

func (manager *defaultDocumentManager) FindBy(query interface{}, returnValues interface{}) error {
	Type := reflect.Indirect(reflect.ValueOf(returnValues)).Type().Elem()
	meta, ok := manager.metadatas[Type]
	if !ok {
		return ErrDocumentNotRegistered
	}
	if err := manager.database.C(meta.collectionName).Find(query).All(returnValues); err != nil {
		return err
	}
	return manager.resolveAllRelations(returnValues)
}

func (manager *defaultDocumentManager) FindAll(returnValues interface{}) error {
	Type := reflect.Indirect(reflect.ValueOf(returnValues)).Type().Elem()
	meta, ok := manager.metadatas[Type]
	if !ok {
		return ErrDocumentNotRegistered
	}
	if err := manager.database.C(meta.collectionName).Find(nil).All(returnValues); err != nil {
		return err
	}
	return manager.resolveAllRelations(returnValues)
}

func (manager *defaultDocumentManager) FindOne(query interface{}, returnValue interface{}) error {
	meta, ok := manager.metadatas[reflect.TypeOf(returnValue)]
	if !ok {
		return ErrDocumentNotRegistered
	}
	if err := manager.database.C(meta.collectionName).Find(query).One(returnValue); err != nil {
		return err
	}
	return manager.resolveRelations(returnValue)
}
func (manager *defaultDocumentManager) FindID(documentID interface{}, returnValue interface{}) error {
	meta, ok := manager.metadatas[reflect.TypeOf(returnValue)]
	if !ok {
		return ErrDocumentNotRegistered
	}
	err := manager.database.C(meta.collectionName).FindId(documentID).One(returnValue)
	if err != nil {
		return err
	}
	return manager.resolveRelations(returnValue)
}
func (manager *defaultDocumentManager) resolveAllRelations(values interface{}) error {
	return nil
}

// resolveRelations resolves all relations for value if value's type was registered
// with the document manager
func (manager *defaultDocumentManager) resolveRelations(value interface{}) error {
	manager.log("Resolving relation for :", reflect.TypeOf(value))
	meta, ok := manager.metadatas[reflect.TypeOf(value)]
	if !ok {
		return ErrDocumentNotRegistered
	}
	if meta.hasRelation() {
		Value := reflect.ValueOf(value)
		fields := meta.getFieldsWithRelation()
		for _, field := range fields {
			switch field.relation.relation {
			case referenceMany:
				fetchedIds := map[string]interface{}{}
				idValue, err := manager.metadatas.GetIDForValue(value)
				if err != nil {
					return err
				}
				err = manager.GetDB().C(meta.collectionName).FindId(idValue).Select(bson.M{field.key: 1}).One(&fetchedIds)
				if err != nil {
					return err
				}
				ids := fetchedIds[field.key].([]interface{})
				relatedMeta, Type := manager.metadatas.FindMetadataByCollectionName(field.relation.targetDocument)
				if Type == nil {
					return ErrDocumentNotRegistered
				}
				relatedResults := reflect.New(reflect.SliceOf(Type))
				err = manager.GetDB().C(relatedMeta.collectionName).Find(bson.M{"_id": bson.M{"$in": ids}}).All(relatedResults.Interface())
				if err != mgo.ErrNotFound && err != nil {
					return err
				}
				Value.Elem().FieldByName(field.name).Set(relatedResults.Elem())
			case referenceOne:
				// we need to requery the main document to get the id of the related document
				document := map[string]interface{}{}
				idValue, err := manager.metadatas.GetIDForValue(value)
				if err != nil {
					return err
				}
				err = manager.GetDB().C(meta.collectionName).FindId(idValue).Select(bson.M{field.key: 1}).One(&document)
				if err != nil {
					return err
				}
				if document[field.key] == nil {
					// no referenced document was found
					continue
				}
				id := document[field.key].(bson.ObjectId)
				relatedMeta, Type := manager.metadatas.FindMetadataByCollectionName(field.relation.targetDocument)
				if Type == nil {
					return ErrDocumentNotRegistered
				}
				relatedResult := reflect.New(Type)

				err = manager.GetDB().C(relatedMeta.collectionName).FindId(id).One(relatedResult.Interface())
				if err != mgo.ErrNotFound && err != nil {
					return err
				}
				Value.Elem().FieldByName(field.name).Set(relatedResult.Elem())
			}
		}
	}
	return nil
}
func stripID(Map map[string]interface{}) map[string]interface{} {
	delete(Map, "_id")
	return Map
}
func isZero(value interface{}) bool {
	Value := reflect.ValueOf(value)
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
	result := "\rmetadata : {"
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
