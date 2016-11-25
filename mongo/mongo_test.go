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

// In all applications, users have a role
type Role struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Title string
}

// Now define the BlogPost document:
type User struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Name  string
	Email string
	Posts []*Post `odm:"referenceMany(targetDocument:Post,cascade:all)"`
	Role  *Role   `odm:"referenceOne(targetDocument:Role,cascade:Persist)"`
}

func TestDocumentManager_Persist(t *testing.T) {
	dm, done := GetDocumentManager(t)
	defer done()
	user := &User{Name: "John", Email: "john@example.com", ID: bson.NewObjectId()}
	post := &Post{Title: "First Post Title", Body: "First Post Body", Created: time.Now()}
	role := &Role{Title: "Editor"}
	user.Posts = append(user.Posts, post)
	user.Role = role
	dm.Persist(user)
	err := dm.Flush()
	test.Fatal(t, err, nil)
	SubTestDocumentManager_FindOne(dm, t)
}

func SubTestDocumentManager_FindOne(dm mongo.DocumentManager, t *testing.T) {
	user := new(User)
	err := dm.FindOne(bson.M{"name": "John"}, user)
	test.Fatal(t, err, nil)
	test.Fatal(t, user.Role != nil, true)
	test.Fatal(t, user.Role.Title, "Editor")
	SubTestDocumentManager_FindID(user.ID, dm, t)
	SubTestDocumentManager_Remove(user.ID, dm, t)
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

func GetDocumentManager(t *testing.T) (dm mongo.DocumentManager, done func()) {
	session, err := mgo.Dial(os.Getenv("MONGODB_TEST_SERVER"))
	test.Fatal(t, err, nil)
	dm = mongo.NewDocumentManager(session.DB("feedpress_test"))
	dm.SetLogger(test.NewTestLogger(t))
	dm.Register("User", new(User))
	dm.Register("Post", new(Post))
	dm.Register("Role", new(Role))
	done = func() {
		session.DB("feedpress_test").C("User").DropCollection()
		session.DB("feedpress_test").C("Post").DropCollection()
		session.DB("feedpress_test").C("Role").DropCollection()
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
