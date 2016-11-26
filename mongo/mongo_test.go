package mongo_test

import (
	"os"
	"testing"
	"time"

	"github.com/Mparaiso/go-tiger/mongo"
	"github.com/Mparaiso/go-tiger/test"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	Posts []*Post `odm:"referenceMany(targetDocument:Post,cascade:all)"`    // cascade changes on persist AND remove
	Role  *Role   `odm:"referenceOne(targetDocument:Role,cascade:Persist)"` // cascade changes only on persist
}

func TestDocumentManager_Persist(t *testing.T) {
	t.Skip()
	dm, done := GetDocumentManager(t)
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
	Projects []*Project    `odm:"referenceMany(targetDocument:Project,mappedBy:Employee)"`
}

type Project struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Title    string        `bson:"Title"`
	Employee *Employee     `bson:"Employee" odm:"referenceOne(targetDocument:Employee)"`
	Client   *Client       `odm:"referenceOne(targetDocument:Client,mappedBy:Projects)"`
}

func TestMappedBy(t *testing.T) {

	dm, done := GetDocumentManager(t)
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
}

func GetDocumentManager(t *testing.T) (dm mongo.DocumentManager, done func()) {
	session, err := mgo.Dial(os.Getenv("MONGODB_TEST_SERVER"))

	test.Fatal(t, err, nil)
	// mgo.SetLogger(MongoLogger{t})
	// mgo.SetDebug(true)
	dm = mongo.NewDocumentManager(session.DB("feedpress_test"))
	dm.SetLogger(test.NewTestLogger(t))

	done = func() {
		session.DB("feedpress_test").C("User").DropCollection()
		session.DB("feedpress_test").C("Post").DropCollection()
		session.DB("feedpress_test").C("Role").DropCollection()
		session.DB("feedpress_test").C("Employee").DropCollection()
		session.DB("feedpress_test").C("Project").DropCollection()
		session.Close()
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

// Given a mongodb server
// 		Given a collection on the server
//		When a document A is persisted
//		It should not return an error
// 		When the document A is fetched
//		It should return the correct Document
func TestMongo(t *testing.T) {
	t.Skip()
	type Test struct {
		ID bson.ObjectId `bson:"_id,omitempty"`
		Name,
		Description string
	}

	session, err := mgo.Dial(os.Getenv("MONGODB_TEST_SERVER"))
	test.Fatal(t, err, nil)
	defer session.Close()
	defer session.DB("feedpress_test").C("mongo_tests").DropCollection()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	collection := session.DB("feedpress_test").C("mongo_tests")
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
