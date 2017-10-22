package users

import (
	"gopkg.in/mgo.v2"
	"os"
	"testing"
)

func TestMongoStore(t *testing.T) {
	mongoAddr := os.Getenv("MONGOADDR")
	if len(mongoAddr) == 0 {
		mongoAddr = "localhost:27017"
	}

	// Create a Mongo session.
	session, err := mgo.Dial(mongoAddr)
	if err != nil {
		t.Fatalf("error dialing mongo: %v", err)
	}

	store := NewMongoStore(session, "test", "user")

	// Create a NewUser for testing purpose.
	nu := CreateNewUser()

	// testUser is never stored in the database.
	testUser, err := nu.ToUser()
	if err != nil {
		t.Errorf("error converting NewUser to User: %v", err)
	}

	// Test retrieving user data that was never stored.
	_, err = store.GetByID(testUser.ID)
	if err == nil {
		t.Error("error expected: retrieving data that was never stored")
	}

	_, err = store.GetByEmail(testUser.Email)
	if err == nil {
		t.Error("error expected: retrieving data that was never stored")
	}

	_, err = store.GetByUserName(testUser.UserName)
	if err == nil {
		t.Error("error expected: retrieving data that was never stored")
	}

	// Test inserting a new user into MongoDB.
	user1, err := store.Insert(nu)
	if err != nil {
		t.Errorf("error inserting a new user to MongoDB: %v", err)
	}

	// Test retrieving user data.
	user2, err := store.GetByID(user1.ID)
	if err != nil {
		t.Errorf("error retrieving user data: %v", err)
	}

	if user2.ID != user1.ID {
		t.Errorf("unmatched user ID\ngot: %s\nwant: %s", user2.ID, user1.ID)
	}

	user2, err = store.GetByEmail(user1.Email)
	if err != nil {
		t.Errorf("error retrieving user data: %v", err)
	}

	if user2.Email != user1.Email {
		t.Errorf("unmatched user ID\ngot: %s\nwant: %s", user2.Email, user1.Email)
	}

	user2, err = store.GetByUserName(user1.UserName)
	if err != nil {
		t.Errorf("error retrieving user data: %v", err)
	}

	if user2.UserName != user1.UserName {
		t.Errorf("unmatched user ID\ngot: %s\nwant: %s", user2.UserName, user1.UserName)
	}

	updates := &Updates{
		FirstName: "Foo",
		LastName:  "Bar",
	}

	// Test updating user data.
	err = store.Update(user1.ID, nil)
	if err == nil {
		t.Errorf("error expected: no Updates provided")
	}

	err = store.Update(user1.ID, updates)
	if err != nil {
		t.Errorf("error updating user data: %v", err)
	}

	// Test updated data.
	user3, err := store.GetByID(user1.ID)
	if err != nil {
		t.Errorf("error retrieving user data: %v", err)
	}

	if user3.FirstName != updates.FirstName {
		t.Errorf("FirstName field is not updated\ngot: %s\nwant: %s", user3.FirstName, updates.FirstName)
	}

	if user3.LastName != updates.LastName {
		t.Errorf("FirstName field is not updated\ngot: %s\nwant: %s", user3.LastName, updates.LastName)
	}

	// Test deleting user data.
	err = store.Delete(user1.ID)
	if err != nil {
		t.Errorf("error deleting user data: %v", err)
	}

	// Try retrieving the deleted data.
	_, err = store.GetByID(user1.ID)
	if err == nil {
		t.Error("error expected: retrieving data that has been deleted")
	}
}
