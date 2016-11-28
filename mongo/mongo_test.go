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
	"fmt"
	"log"
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
	Posts []*Post `odm:"referenceMany(targetDocument:Post,cascade:all)"`    // cascade changes on persist AND remove
	Role  *Role   `odm:"referenceOne(targetDocument:Role,cascade:Persist)"` // cascade changes only on persist
}

func TestDocumentManager_Persist(t *testing.T) {
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
	Projects []*Project    `odm:"referenceMany(targetDocument:Project,mappedBy:Employee,load:eager)"`
}

type Project struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Title    string        `bson:"Title"`
	Employee *Employee     `odm:"referenceOne(targetDocument:Employee,load:eager)"`
	Client   *Client       `odm:"referenceOne(targetDocument:Client,mappedBy:Projects,load:eager)"`
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
	projects := []*Project{}
	err = dm.FindAll(&projects)
	test.Fatal(t, err, nil)
	// test resolveAllRelations for Employee's Projects : referenceMany(mappedBy:Projects)
	test.Fatal(t, len(projects[0].Employee.Projects), 2)
}

type Author struct {
	// ID is the mongo id, needs to be set explicitly
	ID bson.ObjectId `bson:"_id,omitempty"`

	// Name is the author's name
	// unfortunatly mgo driver lowercases mongo keys by default
	// so always explicitely set the key name to the field name
	Name string `bson:"Name"`

	// Articles written by Author , the inverse owning side of the Article/Author relationship
	// Author document in the db WiLL not reference Articles directly BUT loading an Author from the db
	// will also fetch Articles with the related Author.
	Articles []*Article `odm:"referenceMany(targetDocument:Article,mappedBy:Author)"`
}

type Tag struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `bson:"Name"`
	// Articles that have the tag.
	// Although Tag doesn't not hold a reference to articles in the database,
	// related articles are loaded when Tag is fetched.
	Articles []*Article `odm:"referenceMany(targetDocument:Article,mappedBy:Tags)"`
}

type Article struct {
	ID bson.ObjectId `bson:"_id,omitempty"`

	// The title of the blog post
	Title string `bson:"Title"`

	// The author of the post, the owning side of the Article/Author relationship
	Author *Author `odm:"referenceOne(targetDocument:Author)"`

	// Article references many tags.
	// cascade tells the document manager to automatically persist related Tags when Article is persisted,
	// it will also automatically remove related tags as well.
	Tags []*Tag `odm:"referenceMany(targetDocument:Tag,cascade:all)"`
}

func Example() {

	/* we previously defined the following types :

	type Author struct {
		// ID is the mongo id, needs to be set explicitly
		ID bson.ObjectId `bson:"_id,omitempty"`

		// Name is the author's name
		// unfortunatly mgo driver lowercases mongo keys by default
		// so always explicitely set the key name to the field name
		Name string `bson:"Name"`

		// Articles written by Author , the inverse owning side
		// of the Article/Author relationship
		// Author document in the db WiLL not reference Articles directly
		// BUT loading an Author from the db
		// will also fetch Articles with the related Author.
		Articles []*Article `odm:"referenceMany(targetDocument:Article,mappedBy:Author)"`
	}

	type Tag struct {
		ID   bson.ObjectId `bson:"_id,omitempty"`

		Name string        `bson:"Name"`

		// Articles that have the tag.
		// Although Tag doesn't not hold a reference to articles in the database,
		// related articles are loaded when Tag is fetched.
		Articles []*Article `odm:"referenceMany(targetDocument:Article,mappedBy:Tags)"`
	}
	type Article struct {
		ID bson.ObjectId `bson:"_id,omitempty"`

		// The title of the blog post
		Title string `bson:"Title"`

		// The author of the post, the owning side of
		// the Article/Author relationship.
		Author *Author `odm:"referenceOne(targetDocument:Author)"`

		// Article references many tags.
		// cascade tells the document manager to automatically persist related Tags when Article is persisted,
		// it will also automatically remove related tags as well.
		Tags []*Tag `odm:"referenceMany(targetDocument:Tag,cascade:all)"`
	}

	*/

	// create a mongodb connection
	session, err := mgo.Dial(os.Getenv("MONGODB_TEST_SERVER"))
	if err != nil {
		log.Println("error connecting to the db", err)
		return
	}
	// select a database
	db := session.DB(os.Getenv("MONGODB_TEST_DB"))
	defer cleanUp(db)

	// create a document manager
	documentManager := mongo.NewDocumentManager(db)
	documentManager.SetLogger(test.NewTestLogger(&test.ExampleTester{log.New(os.Stderr, "", log.LstdFlags)}))
	// register the types into the document manager
	if err = documentManager.RegisterMany(map[string]interface{}{
		"Article": new(Article),
		"Author":  new(Author),
		"Tag":     new(Tag),
	}); err != nil {
		log.Println("error registering types", err)
		return
	}
	// create some documents
	author := &Author{Name: "John Doe"}
	programmingTag := &Tag{Name: "programming"}
	article1 := &Article{Title: "Go tiger!", Author: author, Tags: []*Tag{{Name: "go"}, programmingTag}}
	article2 := &Article{Title: "MongoDB", Author: author, Tags: []*Tag{programmingTag}}

	// plan for saving
	// we don't need to explicitly persist Tags since
	// Article type cascade persists tags
	documentManager.Persist(author)
	documentManager.Persist(article1)
	documentManager.Persist(article2)

	// do commit changes
	if err = documentManager.Flush(); err != nil {
		log.Println("error saving documents", err)
		return
	}

	/* This is how collections now look like :

	Articles :

	{ 	Title: "MongoDB",
		_id: ObjectId("5839eccc35db821e2c9bc005"),
		author: ObjectId("5839eccc35db821e2c9bc003"),
		tags: [ ObjectId("5839eccc35db821e2c9bc006") ]
	}
	{ 	Title: "Go tiger!",
		_id: ObjectId("5839eccc35db821e2c9bc004"),
		author: ObjectId("5839eccc35db821e2c9bc003"),
		tags: [ ObjectId("5839eccc35db821e2c9bc007"), ObjectId("5839eccc35db821e2c9bc006") ]
	}

	Authors :

	{ Name: "John Doe", _id: ObjectId("5839eccc35db821e2c9bc003") }

	Tags :

	{ Name: "programming", _id: ObjectId("5839eccc35db821e2c9bc006") }
	{ Name: "go", _id: ObjectId("5839eccc35db821e2c9bc007") }
	*/

	// query the database
	author = new(Author)
	if err = documentManager.FindOne(bson.M{"Name": "John Doe"}, author); err != nil {
		log.Println("error fetching author", err)
		return
	}
	fmt.Println("author's name :", author.Name)
	fmt.Println("number of author's articles :", len(author.Articles))

	// query a document by ID
	articleID := author.Articles[0].ID
	article := new(Article)
	if err = documentManager.FindID(articleID, article); err != nil {
		log.Println("error fetching author by id", err)
		return
	}
	fmt.Println("The name of the author of the article :", article.Author.Name)
	fmt.Println("The number of article's tags :", len(article.Tags))

	// fetch all documents
	articles := []*Article{}
	if err = documentManager.FindAll(&articles); err != nil {
		log.Println("error fetching articles", err)
		return
	}
	fmt.Println("articles length :", len(articles))

	// or query specific documents
	articles = []*Article{}
	if err = documentManager.FindBy(bson.M{"Title": "MongoDB"}, &articles); err != nil {
		log.Println("error fetching articles by title", err)
		return
	}
	fmt.Println("articles length :", len(articles))

	// remove a document
	documentManager.Remove(articles[0])
	if err = documentManager.Flush(); err != nil {
		log.Println("error removing article", err)
		return
	}
	// since removing an article remove the related tags, 'programming' tag
	// should not exist in the db

	tag := new(Tag)
	if err = documentManager.FindOne(bson.M{"Name": "programming"}, tag); err != mgo.ErrNotFound {
		log.Printf("%+v\n", tag)
		log.Println("the error should be : mgo.ErrNotFound, got ", err)
		return
	}

	// use complex queries

	author = new(Author)
	// query one document
	query := documentManager.CreateQuery().
		Find(bson.M{"Name": "John Doe"}).Select(bson.M{"Name": 1})

	if err = query.One(author); err != nil {
		log.Println("Error querying one author", err)
		return
	}
	fmt.Println("name:", author.Name)

	authors := []*Author{}
	// query multiple documents
	query = documentManager.CreateQuery().
		Find(bson.M{"Name": bson.M{"$ne": "Jane Doe"}}).
		Select(bson.M{"ID": 1, "Name": 1}).
		Skip(0).Limit(50).Sort("Name")

	if err = query.All(&authors); err != nil {
		log.Println("Error querying authors", err)
		return
	}
	fmt.Println("authors:", len(authors))
	// Output:
	// author's name : John Doe
	// number of author's articles : 2
	// The name of the author of the article : John Doe
	// The number of article's tags : 2
	// articles length : 2
	// articles length : 1
	// name: John Doe
	// authors: 1
}

func cleanUp(db *mgo.Database) {
	for _, collection := range []string{"Article", "Tag", "Author"} {
		db.C(collection).DropCollection()
	}
	db.Session.Close()
}

func GetDocumentManager(t *testing.T) (dm mongo.DocumentManager, done func()) {
	session, err := mgo.Dial(os.Getenv("MONGODB_TEST_SERVER"))
	test.Fatal(t, err, nil)
	dm = mongo.NewDocumentManager(session.DB(os.Getenv("MONGODB_TEST_DB")))
	err = dm.GetDB().DropDatabase()
	test.Fatal(t, err, nil)
	if debug == true {
		mgo.SetLogger(MongoLogger{t})
		mgo.SetDebug(true)
		dm.SetLogger(test.NewTestLogger(t))
	}
	done = func() {
		dm.GetDB().DropDatabase()
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
