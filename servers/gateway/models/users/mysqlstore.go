package users

import (
	"database/sql"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

// Various SQL statements we will need to execute.

// SQL to select a particular user by ID.
// Use `?` for column values that we will get at runtime.
const sqlSelectUserByID = `select * from user where id=?`

// SQL to select a particular user by email.
const sqlSelectUserByEmail = `select * from user where email=?`

// SQL to select a particular user by username.
const sqlSelectUserByUserName = `select * from user where username=?`

// SQL to insert a new user row.
const sqlInsertUser = `insert into user(id,email,passhash,username,firstname,lastname,photourl) values (?,?,?,?,?,?,?)`

// SQL to update user.
const sqlUpdate = `update user set firstname=?, lastname=? where id=?`

// SQL to delete user.
const sqlDelete = `delete from user where id=?`

type userRow struct {
	id        string
	email     string
	passhash  []byte
	username  string
	firstname string
	lastname  string
	photourl  string
}

// MySQLStore implements Store for a MySQL database.
type MySQLStore struct {
	// a live reference to the database.
	db *sql.DB
}

// NewMySQLStore constructs a MySQLStore.
func NewMySQLStore(db *sql.DB) *MySQLStore {
	if db == nil {
		panic("nil pointer passed to NewMySQLStore")
	}

	return &MySQLStore{
		db: db,
	}
}

// GetByID returns the User with the given ID.
func (store *MySQLStore) GetByID(id bson.ObjectId) (*User, error) {
	rows, err := store.db.Query(sqlSelectUserByID, id.Hex())
	if err != nil {
		return nil, fmt.Errorf("error selecting user: %v", err)
	}

	users, err := scanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning user: %s", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found")
	}

	// Return the first (and only) element from the slice.
	return users[0], nil
}

// GetByEmail returns the User with the given email.
func (store *MySQLStore) GetByEmail(email string) (*User, error) {
	rows, err := store.db.Query(sqlSelectUserByEmail, email)
	if err != nil {
		return nil, fmt.Errorf("error selecting user: %v", err)
	}

	users, err := scanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning user: %s", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found")
	}

	// Return the first (and only) element from the slice.
	return users[0], nil
}

// GetByUserName returns the User with the given Username.
func (store *MySQLStore) GetByUserName(username string) (*User, error) {
	rows, err := store.db.Query(sqlSelectUserByUserName, username)
	if err != nil {
		return nil, fmt.Errorf("error selecting user: %v", err)
	}

	users, err := scanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning user: %s", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found")
	}

	// Return the first (and only) element from the slice.
	return users[0], nil
}

// Insert converts the NewUser to a User, inserts
// it into the database, and returns it.
func (store *MySQLStore) Insert(newUser *NewUser) (*User, error) {
	user, err := newUser.ToUser()
	if err != nil {
		return nil, fmt.Errorf("error converting NewUser to User: %v", err)
	}

	// Use transaction to make sure inserts to be atomic (all or nothing).
	tx, err := store.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %v", err)
	}

	// Execute the insert to the `user` table.
	// The .Hex() method of bson.ObjectId will return
	// the hexadecimal string representation of the binary
	// object ID, which is human-readable.
	_, err = tx.Exec(sqlInsertUser, user.ID.Hex(), user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	if err != nil {
		// Rollback the transaction if there's an error.
		tx.Rollback()
		return nil, fmt.Errorf("error inserting user: %v", err)
	}

	// Now commit the transaction so that all those inserts are atomic.
	err = tx.Commit()
	if err != nil {
		// Try to rollback if we can't commit
		tx.Rollback()
		return nil, fmt.Errorf("error committing insert transaction: %v", err)
	}

	return user, nil
}

// Update applies UserUpdates to the given user ID.
func (store *MySQLStore) Update(userID bson.ObjectId, updates *Updates) error {
	if updates == nil {
		return fmt.Errorf("Updates is nil")
	}

	_, err := store.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	_, err = store.db.Exec(sqlUpdate, updates.FirstName, updates.LastName, userID.Hex())
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	return nil
}

// Delete deletes the user with the given ID.
func (store *MySQLStore) Delete(userID bson.ObjectId) error {

	// Check if the user exists on database.
	_, err := store.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	_, err = store.db.Exec(sqlDelete, userID.Hex())
	if err != nil {
		return fmt.Errorf("error deleting data: %v", err)
	}

	return nil
}

// scanUsers scans query result rows into a []*User.
func scanUsers(rows *sql.Rows) ([]*User, error) {
	// Ensure the rows are closed regardless of how
	// we exit this function.
	defer rows.Close()

	// Empty slice of *User to hold the result.
	users := []*User{}

	// Create a userRow struct to scan the data into.
	row := userRow{}

	for rows.Next() {
		// Scan each record into User struct.
		err := rows.Scan(&row.id, &row.email, &row.passhash, &row.username, &row.firstname, &row.lastname, &row.photourl)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		user := &User{
			ID:        bson.ObjectIdHex(row.id),
			Email:     row.email,
			PassHash:  row.passhash,
			UserName:  row.username,
			FirstName: row.firstname,
			LastName:  row.lastname,
			PhotoURL:  row.photourl,
		}

		users = append(users, user)
	}

	// If there was an error reading rows off the network
	// rows.Err() will return a non-nil value.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %v", err)
	}

	return users, nil
}
