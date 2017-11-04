package users

import (
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestMemStore(t *testing.T) {

	store := NewMemStore()

	// Create a NewUser for testing purpose.
	nu := CreateNewUser()

	// testUser is never stored in the database.
	testUser, err := nu.ToUser()
	if err != nil {
		t.Errorf("error converting NewUser to User: %s", err)
	}

	// Test retrieving user data that was never stored.
	_, err = store.GetByID(testUser.ID)
	if err == nil {
		t.Error("expected error when attempting to retrieve user data that was never stored")
	}

	_, err = store.GetByEmail(testUser.Email)
	if err == nil {
		t.Error("expected error when attempting to retrieve user data that was never stored")
	}

	_, err = store.GetByUserName(testUser.UserName)
	if err == nil {
		t.Error("expected error when attempting to retrieve user data that was never stored")
	}

	// Test inserting a new user.
	user1, err := store.Insert(nu)
	if err != nil {
		t.Errorf("error inserting a new user to MemStore: %s", err)
	}

	// Test retrieving user data.
	user2, err := store.GetByID(user1.ID)
	if err != nil {
		t.Errorf("error retrieving user data: %s", err)
	}

	if user2.ID != user1.ID {
		t.Errorf("unmatched user ID\ngot: %s\nwant: %s", user2.ID, user1.ID)
	}

	user2, err = store.GetByEmail(user1.Email)
	if err != nil {
		t.Errorf("error retrieving user data: %s", err)
	}

	if user2.Email != user1.Email {
		t.Errorf("unmatched user email\ngot: %s\nwant: %s", user2.Email, user1.Email)
	}

	user2, err = store.GetByUserName(user1.UserName)
	if err != nil {
		t.Errorf("error retrieving user data: %s", err)
	}

	if user2.UserName != user1.UserName {
		t.Errorf("unmatched username\ngot: %s\nwant: %s", user2.UserName, user1.UserName)
	}

	updates := &Updates{
		FirstName: "Foo",
		LastName:  "Bar",
	}

	// Test updating user data.
	err = store.Update(user1.ID, nil)
	if err == nil {
		t.Error("expected error when attempting to update user with nil Updates")
	}

	invalidUserID := bson.NewObjectId()
	err = store.Update(invalidUserID, updates)
	if err == nil {
		t.Error("expected error when attempting to update user with invalid user ID")
	}

	err = store.Update(user1.ID, updates)
	if err != nil {
		t.Errorf("error updating user: %s", err)
	}

	// Test updated data.
	user3, err := store.GetByID(user1.ID)
	if err != nil {
		t.Errorf("error retrieving user data: %s", err)
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
		t.Errorf("error deleting user: %s", err)
	}

	// Try retrieving the deleted data.
	_, err = store.GetByID(user1.ID)
	if err == nil {
		t.Error("expected error when attempting to retrieve user data that has already been deleted")
	}

	// Try deleting the same user data again.
	err = store.Delete(user1.ID)
	if err == nil {
		t.Error("expected error when attempting to delete user data that has already been deleted")
	}
}
