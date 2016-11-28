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

package mongo_test

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/Mparaiso/go-tiger/mongo"
	"github.com/Mparaiso/go-tiger/test"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	debug = false
)

type Post struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Title   string
	Body    string
	Created time.Time
}

type Role struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Title string
	Users []*User
}

type User struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Name  string
	Email string
	Posts []*Post `odm:"referenceMany(targetDocument:Post,cascade:all,storeId:PostIds)"` // cascade changes on persist AND remove
	Role  *Role   `odm:"referenceOne(targetDocument:Role,cascade:Persist)"`              // cascade changes only on persist
}

func TestDocumentManager_Persist(t *testing.T) {
	dm, done := getDocumentManager(t)
	defer done()
	dm.Register("User", new(User))
	dm.Register("Post", new(Post))
	dm.Register("Role", new(Role))
	user := &User{Name: "John", Email: "john@example.com", ID: bson.NewObjectId()}
	post := &Post{Title: "First Post Title", Body: "First Post Body", Created: time.Now()}
	role := &Role{Title: "Editor"}
	user.Posts = append(user.Posts, post)
	user.Role = role
	dm.Persist(user)
	err := dm.Flush()
	test.Fatal(t, err, nil)
	id := SubTestDocumentManager_FindOne(dm, t)
	SubTestDocumentManager_FindID(id, dm, t)
	SubTestDocumentManager_FindAll(dm, t)
	SubTestDocumentManager_Remove(id, dm, t)
}

func SubTestDocumentManager_FindOne(dm mongo.DocumentManager, t *testing.T) bson.ObjectId {
	user := new(User)
	err := dm.FindOne(bson.M{"name": "John"}, user)
	test.Fatal(t, err, nil)
	test.Fatal(t, user.Role != nil, true)
	test.Fatal(t, user.Role.Title, "Editor")
	return user.ID
}

func SubTestDocumentManager_FindAll(dm mongo.DocumentManager, t *testing.T) {
	users := []*User{}
	err := dm.FindAll(&users)
	test.Fatal(t, err, nil)
	test.Fatal(t, len(users), 1)
	test.Fatal(t, len(users[0].Posts), 1)
	test.Fatal(t, users[0].Posts[0].Title, "First Post Title")
	test.Fatal(t, users[0].Role != nil, true)
}

func SubTestDocumentManager_FindID(id bson.ObjectId, dm mongo.DocumentManager, t *testing.T) {
	user := new(User)
	err := dm.FindID(id, user)
	test.Fatal(t, err, nil)
	test.Fatal(t, user.ID, id)
}

func SubTestDocumentManager_Remove(id bson.ObjectId, dm mongo.DocumentManager, t *testing.T) {
	user := new(User)
	err := dm.FindID(id, user)
	test.Fatal(t, err, nil)
	test.Fatal(t, len(user.Posts), 1)
	postID := user.Posts[0].ID
	roleID := user.Role.ID
	dm.Remove(user)
	dm.Flush()
	err = dm.FindID(id, user)
	test.Fatal(t, err, mgo.ErrNotFound)
	post := new(Post)
	err = dm.FindID(postID, post)
	test.Fatal(t, err, mgo.ErrNotFound)
	role := new(Role)
	err = dm.FindID(roleID, role)
	test.Fatal(t, err, nil)
}

type Client struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Name     string        `bson:"Name"`
	Projects []*Project    `bson:"Projects,omitempty" odm:"referenceMany(targetDocument:Project)"`
}

type Employee struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Name     string        `bson:"Name"`
	Projects []*Project    `odm:"referenceMany(targetDocument:Project,mappedBy:Employee,load:eager)"`
}

type Project struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Title    string        `bson:"Title"`
	Employee *Employee     `odm:"referenceOne(targetDocument:Employee,load:eager)"`
	Client   *Client       `odm:"referenceOne(targetDocument:Client,mappedBy:Projects,load:eager)"`
}

func TestDocumentManager_Register_StoreIdAnnotation(t *testing.T) {

	type Person struct {
		ID   bson.ObjectId `bson:"_id"`
		Name string        `bson:"Name"`
	}
	type Family struct {
		ID        bson.ObjectId   `bson:"_id"`
		Name      string          `bson:"Name"`
		Members   []*Person       `odm:"referenceMany(targetDocument:Person,storeId:MemberIDs,cascade:all)"`
		MemberIDs []bson.ObjectId `bson:"MemberIDs"`
	}
	dm, done := getDocumentManager(t)
	defer done()
	err := dm.RegisterMany(map[string]interface{}{
		"Person": new(Person),
		"Family": new(Family),
	})
	test.Fatal(t, err, nil)
	family := &Family{Name: "Doe", Members: []*Person{{Name: "John"}, {Name: "Jane"}, {Name: "Jack"}}}
	dm.Persist(family)
	err = dm.Flush()
	test.Fatal(t, err, nil)
	family = new(Family)
	err = dm.FindOne(bson.M{"Name": "Doe"}, family)
	test.Fatal(t, err, nil)
	test.Fatal(t, len(family.MemberIDs), 3)

}

func TestDocumentManager_Register_InvalidAnnotation(t *testing.T) {
	type City struct {
		Name string `odm:"invalidAnnotation"`
	}
	dm := mongo.NewDocumentManager(nil)
	err := dm.Register("City", new(City))
	test.Fatal(t, err, mongo.ErrInvalidAnnotation)
}

func TestDocumentManager_FindAll_MappedBy(t *testing.T) {
	dm, done := getDocumentManager(t)
	defer done()
	err := dm.Register("Employee", new(Employee))
	test.Fatal(t, err, nil)
	err = dm.Register("Project", new(Project))
	test.Fatal(t, err, nil)
	err = dm.Register("Client", new(Client))
	test.Fatal(t, err, nil)
	employee := &Employee{ID: bson.NewObjectId(), Name: "John"}
	project1 := &Project{Title: "First project", Employee: employee}
	project2 := &Project{Title: "Second project", Employee: employee}
	client1 := &Client{Name: "Example", Projects: []*Project{project1, project2}}
	client2 := &Client{Name: "Acme"}
	dm.Persist(employee)
	dm.Persist(project1)
	dm.Persist(project2)
	dm.Persist(client1)
	dm.Persist(client2)
	err = dm.Flush()
	test.Fatal(t, err, nil)
	employee1 := new(Employee)
	err = dm.FindOne(bson.M{"Name": "John"}, employee1)
	test.Fatal(t, err, nil)
	test.Fatal(t, len(employee1.Projects), 2)
	test.Fatal(t, employee1.Projects[0].Employee, employee1)
	test.Fatal(t, employee1.Projects[0].Client != nil, true)
	test.Fatal(t, employee1.Projects[0].Client.Name, "Example")
	test.Fatal(t, employee1.Projects[1].Client.Name, "Example")
	projects := []*Project{}
	err = dm.FindAll(&projects)
	test.Fatal(t, err, nil)
	// test resolveAllRelations for Employee's Projects : referenceMany(mappedBy:Projects)
	test.Fatal(t, len(projects[0].Employee.Projects), 2)
}

func cleanUp(db *mgo.Database) {
	for _, collection := range []string{"Article", "Tag", "Author"} {
		db.C(collection).DropCollection()
	}
	db.Session.Close()
}

func getDB(t *testing.T) *mgo.Database {
	session, err := mgo.Dial(os.Getenv("MONGODB_TEST_SERVER"))
	test.Fatal(t, err, nil)
	if debug == true {
		mgo.SetLogger(MongoLogger{t})
		mgo.SetDebug(true)
	}
	return session.DB(os.Getenv("MONGODB_TEST_DB"))
}

func getDocumentManager(t *testing.T) (dm mongo.DocumentManager, done func()) {

	dm = mongo.NewDocumentManager(getDB(t))
	err := dm.GetDB().DropDatabase()
	test.Fatal(t, err, nil)
	if debug == true {
		dm.SetLogger(test.NewTestLogger(t))
	}
	done = func() {
		dm.GetDB().DropDatabase()
		dm.GetDB().Session.Close()
	}
	return
}

type MongoLogger struct {
	t test.Tester
}

func (m MongoLogger) Output(i int, s string) error {
	m.t.Logf("callstack %d \n output : %s", i, s)
	return nil
}

func TestMongo(t *testing.T) {

	type Test struct {
		ID bson.ObjectId `bson:"_id,omitempty"`
		Name,
		Description string
	}

	session, err := mgo.Dial(os.Getenv("MONGODB_TEST_SERVER"))
	test.Fatal(t, err, nil)
	defer session.Close()
	defer session.DB(os.Getenv("MONGODB_TEST_DB")).C("mongo_tests").DropCollection()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	collection := session.DB(os.Getenv("MONGODB_TEST_DB")).C("mongo_tests")
	test1 := &Test{Name: "Initial", Description: "A simple test"}
	err = collection.Insert(test1)
	test.Fatal(t, err, nil)
	result := new(Test)
	err = collection.Find(bson.M{"name": test1.Name}).One(result)
	test.Error(t, err, nil)
	test.Error(t, result.Description, test1.Description)
	result1 := new(Test)
	err = collection.FindId(result.ID).One(result1)
	test.Error(t, err, nil)
	test.Error(t, result1.ID, result.ID)
}

// TestMain
// test options : -debug
func TestMain(m *testing.M) {
	type Arguments struct {
		Debug bool
	}
	args := Arguments{}
	flag.BoolVar(&args.Debug, "debug", false, "run tests in debug mode")
	flag.Parse()
	debug = args.Debug
	m.Run()
}
