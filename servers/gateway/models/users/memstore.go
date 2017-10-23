package users

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

// MemStore represents a fake store
// that temporarily saves user data in memory.
// It is just a slice of Users stored in memory.
type MemStore struct {
	entries []*User
}

// NewMemStore constructs and returns a new MemStore.
func NewMemStore() *MemStore {
	return &MemStore{
		entries: []*User{},
	}
}

// GetByID returns the User with the given ID from the in-memory store.
func (ms *MemStore) GetByID(id bson.ObjectId) (*User, error) {
	// Search through all Users stored in the in-memory store.
	for _, user := range ms.entries {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// GetByEmail returns the User with the given email from the in-memory store.
func (ms *MemStore) GetByEmail(email string) (*User, error) {
	for _, user := range ms.entries {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// GetByUserName returns the User with the given Username from the in-memory store.
func (ms *MemStore) GetByUserName(username string) (*User, error) {
	for _, user := range ms.entries {
		if user.UserName == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// Insert converts the NewUser to a User, inserts
// it into the in-memory store, and returns it.
func (ms *MemStore) Insert(newUser *NewUser) (*User, error) {
	user, err := newUser.ToUser()
	if err != nil {
		return nil, fmt.Errorf("error converting NewUser to User: %v", err)
	}

	ms.entries = append(ms.entries, user)

	return user, nil
}

// Update applies UserUpdates to the given user ID.
func (ms *MemStore) Update(userID bson.ObjectId, updates *Updates) error {
	user, err := ms.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error retrieving user data")
	}

	user.FirstName = updates.FirstName
	user.LastName = updates.LastName

	return nil
}

// Delete deletes the user with the given ID.
func (ms *MemStore) Delete(userID bson.ObjectId) error {
	for i, user := range ms.entries {
		if user.ID == userID {
			ms.entries = append(ms.entries[:i], ms.entries[i+1:]...)
		}
		return nil
	}

	return fmt.Errorf("error deleting data")
}
