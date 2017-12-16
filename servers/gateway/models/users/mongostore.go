package users

import (
	"fmt"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/indexes"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MongoStore implements Store for MongoDB.
type MongoStore struct {
	session *mgo.Session
	dbname  string
	colname string // collection name
}

// NewMongoStore constructs a new MongoStore.
func NewMongoStore(session *mgo.Session, dbName string, collectionName string) *MongoStore {
	if session == nil {
		panic("nil pointer passed for session")
	}
	return &MongoStore{
		session: session,
		dbname:  dbName,
		colname: collectionName,
	}
}

// GetByID returns the User with the given ID.
func (store *MongoStore) GetByID(id bson.ObjectId) (*User, error) {
	// Create an empty User struct to hold user data retrieved from MongoDB.
	user := &User{}
	err := store.session.DB(store.dbname).C(store.colname).FindId(id).One(user)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from MongoDB: %v", err)
	}
	return user, nil
}

// GetByEmail returns the User with the given email.
func (store *MongoStore) GetByEmail(email string) (*User, error) {
	user := &User{}
	q := bson.M{"email": email}
	err := store.session.DB(store.dbname).C(store.colname).Find(q).One(user)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from MongoDB: %v", err)
	}
	return user, nil
}

// GetByUserName returns the User with the given Username.
func (store *MongoStore) GetByUserName(username string) (*User, error) {
	user := &User{}
	q := bson.M{"username": username}
	err := store.session.DB(store.dbname).C(store.colname).Find(q).One(user)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from MongoDB: %v", err)
	}
	return user, nil
}

// Insert converts the NewUser to a User, inserts
// it into the database, and returns it.
func (store *MongoStore) Insert(newUser *NewUser) (*User, error) {
	user, err := newUser.ToUser()
	if err != nil {
		return nil, fmt.Errorf("error converting NewUser to User: %v", err)
	}

	err = store.session.DB(store.dbname).C(store.colname).Insert(user)
	if err != nil {
		return nil, fmt.Errorf("error inserting data into MongoDB: %v", err)
	}

	return user, nil
}

// Update applies UserUpdates to the given user ID.
func (store *MongoStore) Update(userID bson.ObjectId, updates *Updates) error {
	if updates == nil {
		return fmt.Errorf("Updates is nil")
	}

	change := mgo.Change{
		Update:    bson.M{"$set": updates}, // $set sends a PATCH
		ReturnNew: true,                    // Get back new version rather than old version of the data.
	}
	user := &User{}

	_, err := store.session.DB(store.dbname).C(store.colname).FindId(userID).Apply(change, user)
	if err != nil {
		return fmt.Errorf("error updating MongoDB: %v", err)
	}

	return nil
}

// Delete deletes the user with the given ID.
func (store *MongoStore) Delete(userID bson.ObjectId) error {
	err := store.session.DB(store.dbname).C(store.colname).RemoveId(userID)
	if err != nil {
		return fmt.Errorf("error deleting data: %v", err)
	}

	return nil
}

// Index stores all users email, username, lastname, and firstname into a trie.
func (store *MongoStore) Index() *indexes.Trie {
	user := &User{}
	trie := indexes.NewTrie()

	// Iterate all users from database one at a time.
	iter := store.session.DB(store.dbname).C(store.colname).Find(nil).Iter()

	for iter.Next(user) {
		trie.Insert(user.Email, user.ID)
		trie.Insert(user.UserName, user.ID)
		trie.Insert(user.LastName, user.ID)
		trie.Insert(user.FirstName, user.ID)
	}

	// Report any errors that occurred.
	if err := iter.Err(); err != nil {
		fmt.Printf("error iterating stored documents: %v", err)
	}

	return trie
}

// ConvertToUsers converts all keys(User IDs) in a given map to a slice of User.
func (store *MongoStore) ConvertToUsers(userIDs map[bson.ObjectId]bool) ([]*User, error) {
	users := []*User{}
	for userID := range userIDs {
		user, err := store.GetByID(userID)
		if err != nil {
			return nil, fmt.Errorf("error getting user: %v", err)
		}
		users = append(users, user)
	}

	return users, nil
}
