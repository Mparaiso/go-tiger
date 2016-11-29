mongo-odm
=========

Package mongo-odm provides an object document mapper, or ODM,
 for mongodb and is strongly influenced by Doctrine Mongo ODM.

With mongo-odm, developers no longer need to manually map related documents 
from different collections, odm takes cares of all the busy work automatically
making it easy to model complex data. mongo-odm is written in Go.


#### requirements

	go 1.7
	
#### installation 

	go get github.com/mparaiso/go-tiger/mongo

#### similar projects:

https://github.com/dsmontoya/marango

https://github.com/maxwellhealth/bongo

#### related projects:

https://github.com/search?l=JavaScript&p=2&q=odm&type=Repositories&utf8=%E2%9C%93

https://docs.mongodb.com/ruby-driver/master/tutorials/6.0.0/mongoid-relations/

#### basic usage

```go
	
	package main 
	
	import(
		"github.com/mparaiso/go-tiger/mongo"
		"log"
		"fmt"
		"os"
		mgo "gopkg.in/mgo.v2"
		"gopkg.in/mgo.v2/bson"
	)

	
	type Author struct {
		// ID is the mongo id. it needs to be set explicitly.
		ID bson.ObjectId `bson:"_id,omitempty"`
	
		// Name is the author's name.
		// Unfortunatly mgo driver lowercases mongo keys by default
		// so always explicitely set the key name.
		Name string `bson:"Name"`
	
		// Articles written by Author, the inverse side of the Article/Author relationship
		// Author document does not reference Articles directly in the db 
		// but loading an Author will automatically load related articles.
		Articles []*Article `odm:"referenceMany(targetDocument:Article,mappedBy:Author)"`
	}
	
	type Tag struct {
		ID   bson.ObjectId `bson:"_id,omitempty"`
		
		Name string        `bson:"Name"`
		
		// loading Tag will automatically load the related articles, but Tag 
		// document in the db has no reference to Article.
		Articles []*Article `odm:"referenceMany(targetDocument:Article,mappedBy:Tags)"`
	}
	
	type Article struct {
		ID bson.ObjectId `bson:"_id,omitempty"`
	
		// The title of the blog post
		Title string `bson:"Title"`
	
		// The author of the post, the owning side of the Article/Author relationship
		Author *Author `odm:"referenceOne(targetDocument:Author)"`
	
		// Article references many tags.
		// cascade:all tells the document manager to persist or deleted related Tags 
		// when an article is saved or deleted.
		Tags []*Tag `odm:"referenceMany(targetDocument:Tag,cascade:all)"`
	}
	
	func main() {
		
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
	
		// register types into the document manager
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
		// Article cascade persists tags
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
	
		// remove an article
		documentManager.Remove(articles[0])
		if err = documentManager.Flush(); err != nil {
			log.Println("error removing article", err)
			return
		}
		// since removing an article automatically remove the related tags, 
		// 'programming' tag was removed too.
		
		tag := new(Tag)
		if err = documentManager.FindOne(bson.M{"Name": "programming"}, tag); err != mgo.ErrNotFound {
			log.Printf("%+v\n", tag)
			log.Println("the error should be : mgo.ErrNotFound, got ", err)
			return
		}
	
		// complex queries
	
		
		// query one author, only return his name
		
		author = new(Author)
		query := documentManager.CreateQuery().
			Find(bson.M{"Name": "John Doe"}).Select(bson.M{"Name": 1})
	
		if err = query.One(author); err != nil {
			log.Println("Error querying one author", err)
			return
		}
		fmt.Println("name:", author.Name)
	
		authors := []*Author{}
		
		// query multiple documents with limit, offset ordered by Name
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
	
```