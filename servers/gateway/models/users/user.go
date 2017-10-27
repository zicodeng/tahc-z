package users

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
	"io"
	"net/mail"
	"strings"
)

const gravatarBasePhotoURL = "https://www.gravatar.com/avatar/"

var bcryptCost = 13

// User represents a user account in the database.
type User struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	Email     string        `json:"email"`
	PassHash  []byte        `json:"-"` // Stored, but not encoded to clients.
	UserName  string        `json:"userName"`
	FirstName string        `json:"firstName"`
	LastName  string        `json:"lastName"`
	PhotoURL  string        `json:"photoURL"`
}

// Credentials represents user sign-in credentials.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewUser represents a new user signing up for an account.
type NewUser struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	PasswordConf string `json:"passwordConf"`
	UserName     string `json:"userName"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
}

// Updates represents allowed updates to a user profile.
type Updates struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// Validate validates the new user and returns an error if
// any of the validation rules fail, or nil if its valid.
func (nu *NewUser) Validate() error {

	// Email field must be a valid email address.
	_, err := mail.ParseAddress(nu.Email)
	if err != nil {
		return fmt.Errorf("error parsing email: %v", err)
	}

	// Password must be at least 6 characters.
	if len(nu.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	// Password and PasswordConf must match.
	if nu.Password != nu.PasswordConf {
		return fmt.Errorf("password must match password confirmation")
	}

	// UserName must be non-zero length.
	if len(nu.UserName) == 0 {
		return fmt.Errorf("username must be non-zero length")
	}

	// FirstName must be non-zero length.
	if len(nu.FirstName) == 0 {
		return fmt.Errorf("first name must be non-zero length")
	}

	// LastName must be non-zero length.
	if len(nu.LastName) == 0 {
		return fmt.Errorf("last name must be non-zero length")
	}

	return nil
}

// ToUser converts the NewUser to a User, setting the
// PhotoURL and PassHash fields appropriately.
func (nu *NewUser) ToUser() (*User, error) {

	// Construct a User based on NewUser.
	usr := &User{
		UserName:  nu.UserName,
		FirstName: nu.FirstName,
		LastName:  nu.LastName,
	}

	// Trim leading and trailing whitespace from an email address.
	email := strings.TrimSpace(nu.Email)

	// Force all characters in the email to be lower-case.
	email = strings.ToLower(email)

	// Update Email field.
	usr.Email = email

	// md5 hash the final email string.
	h := md5.New()
	io.WriteString(h, email)
	result := hex.EncodeToString(h.Sum(nil))

	// Set the PhotoURL field of the new User to
	// the Gravatar PhotoURL for the user's email address.
	photoURL := gravatarBasePhotoURL + result
	usr.PhotoURL = photoURL

	// Set the ID field of the new User
	// to a new bson ObjectId.
	usr.ID = bson.NewObjectId()

	// Call .SetPassword() to set the PassHash
	// field of the User to a hash of the NewUser.Password.
	err := usr.SetPassword(nu.Password)
	if err != nil {
		return nil, fmt.Errorf("error setting password hash of the User: %v", err)
	}

	return usr, nil
}

// FullName returns the user's full name, in the form:
// "<FirstName> <LastName>"
// If either first or last name is an empty string, no
// space is put betweeen the names.
func (u *User) FullName() string {
	fullName := ""

	if len(u.FirstName) > 0 {
		fullName += u.FirstName
	}

	if len(u.FirstName) > 0 && len(u.LastName) > 0 {
		fullName += " "
	}

	if len(u.LastName) > 0 {
		fullName += u.LastName
	}

	return fullName
}

// SetPassword hashes the password and stores it in the PassHash field.
func (u *User) SetPassword(password string) error {
	// Automatically generates salt while hashing.
	// second parameter is the adaptive cost factor,
	// which controls the speed at which the algorithm runs.
	// The higher the cost factor, the slower the algorithm runs.
	// It wants the password as a byte slice, so convert using []byte()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("error generating bcrypt hash: %v", err)
	}
	u.PassHash = passwordHash
	return nil
}

// Authenticate compares the plaintext password against the stored hash
// and returns an error if they don't match, or nil if they do.
func (u *User) Authenticate(password string) error {
	err := bcrypt.CompareHashAndPassword(u.PassHash, []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password: %v", err)
	}
	return nil
}

// ApplyUpdates applies the updates to the user. An error
// is returned if the updates are invalid
func (u *User) ApplyUpdates(updates *Updates) error {
	if len(updates.FirstName) == 0 {
		return fmt.Errorf("first name must be non-zero length")
	}

	if len(updates.LastName) == 0 {
		return fmt.Errorf("last name must be non-zero length")
	}

	u.FirstName = updates.FirstName
	u.LastName = updates.LastName

	return nil
}
