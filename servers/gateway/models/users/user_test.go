package users

import (
	"crypto/md5"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"io"
	"reflect"
	"testing"
)

// NewUser creates a NewUser with all valid fields.
func CreateNewUser() *NewUser {
	return &NewUser{
		Email:        "zicodeng@gmail.com",
		Password:     "password",
		PasswordConf: "password",
		UserName:     "zicodeng",
		FirstName:    "Zico",
		LastName:     "Deng",
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name              string
		invalidFieldName  string
		invalidFieldValue string
		hint              string
	}{
		{
			"valid new user",
			"",
			"",
			"this is a valid NewUser, error should be nil",
		},
		{
			"invalid email",
			"Email",
			"invalid",
			"error parsing email",
		},
		{
			"invalid email",
			"Email",
			"invalid@",
			"error parsing email",
		},
		{
			"invalid email",
			"Email",
			"@invalid",
			"error parsing email",
		},
		{
			"invalid password",
			"Password",
			"12345",
			"password must be at least 6 characters",
		},
		{
			"invalid password confirmation",
			"PasswordConf",
			"wordpass",
			"password must match password confirmation",
		},
		{
			"empty username",
			"UserName",
			"",
			"username must be non-zero length",
		},
		{
			"empty first name",
			"FirstName",
			"",
			"first name must be non-zero length",
		},
		{
			"empty last name",
			"LastName",
			"",
			"last name must be non-zero length",
		},
	}

	for _, c := range cases {
		nu := CreateNewUser()

		// Mutate valid fields to invalid so we can test them.
		v := reflect.ValueOf(nu).Elem().FieldByName(c.invalidFieldName)

		if v.IsValid() {
			v.SetString(c.invalidFieldValue)
		}

		err := nu.Validate()

		// Test valid cases.
		if c.invalidFieldName == "" && c.invalidFieldValue == "" {
			if err != nil {
				t.Errorf("\ncase: %s\ninvalid field: {%s: %s}\nwant: %s", c.name, c.invalidFieldName, c.invalidFieldValue, c.hint)
			}
		} else {
			// Test invalid cases.
			if err == nil {
				t.Errorf("\ncase: %s\ninvalid field: {%s: %s}\nwant: %s", c.name, c.invalidFieldName, c.invalidFieldValue, c.hint)
			}
		}
	}
}

func TestToUser(t *testing.T) {
	cases := []struct {
		name         string
		validEmail   string
		invalidEmail string
		hint         string
	}{
		{
			"email contains leading or trailing whitespace",
			"zicodeng@gmail.com",
			" zicodeng@gmail.com   ",
			"make sure to trim leading and trailing whitespace from an email address",
		},
		{
			"email contains uppercase characters",
			"foo@bar.edu",
			"foo@BAR.edu",
			"make sure to convert all characters to lower-case",
		},
	}

	for _, c := range cases {
		nu := CreateNewUser()
		nu.Email = c.invalidEmail

		// Convert NewUser to User.
		usr, err := nu.ToUser()
		if err != nil {
			t.Errorf("error converting NewUser to User\n")
		}

		if usr == nil {
			t.Errorf("ToUser() returned nil\n")
		}

		// Test Email.
		if usr.Email != c.validEmail {
			t.Errorf("\ncase: %s\ngot: %s\nwant: %s\nhint: %s", c.name, usr.Email, c.validEmail, c.hint)
		}

		// Test PhotoURL.
		h := md5.New()
		io.WriteString(h, c.validEmail)
		result := hex.EncodeToString(h.Sum(nil))
		photoURL := gravatarBasePhotoURL + result

		if len(usr.PhotoURL) == 0 {
			t.Errorf("PhotoURL field is empty\n")
		}

		if usr.PhotoURL != photoURL {
			t.Errorf("invalid PhotoURL")
		}

		// Test PassHash.
		if len(usr.PassHash) == 0 {
			t.Errorf("password hash is empty")
		}

		err = bcrypt.CompareHashAndPassword(usr.PassHash, []byte(nu.Password))
		if err != nil {
			t.Errorf("invalid password: %v", err)
		}
	}
}

func TestFullName(t *testing.T) {
	cases := []struct {
		name           string
		firstName      string
		lastName       string
		expectedOutput string
	}{
		{
			"both first name and last name non-empty",
			"Foo",
			"Bar",
			"Foo Bar",
		},
		{
			"first name empty",
			"",
			"Bar",
			"Bar",
		},
		{
			"last name empty",
			"Foo",
			"",
			"Foo",
		},
		{
			"both first name and last name empty",
			"",
			"",
			"",
		},
	}

	for _, c := range cases {
		nu := CreateNewUser()
		nu.FirstName = c.firstName
		nu.LastName = c.lastName

		usr, err := nu.ToUser()
		if err != nil {
			t.Errorf("error converting NewUser to User\n")
		}

		fullName := usr.FullName()
		if fullName != c.expectedOutput {
			t.Errorf("\ncase: %s\ngot: %s\nwant: %s", c.name, fullName, c.expectedOutput)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	usr := &User{}

	if err := usr.SetPassword("password"); err != nil {
		t.Errorf("error setting password: %v", err)
	}

	if err := usr.Authenticate("password"); err != nil {
		t.Errorf("the password is valid, no error expected")
	}

	if err := usr.Authenticate(""); err == nil {
		t.Errorf("empth password")
	}

	if err := usr.Authenticate("wordpass"); err == nil {
		t.Errorf("invalid password")
	}
}

func TestApplyUpdates(t *testing.T) {
	cases := []struct {
		name        string
		updates     *Updates
		expectError bool
	}{
		{
			"valid updates",
			&Updates{
				"Foo",
				"Bar",
			},
			false,
		},
		{
			"first name empty",
			&Updates{
				"",
				"Bar",
			},
			true,
		},
		{
			"last name empty",
			&Updates{
				"Foo",
				"",
			},
			true,
		},
		{
			"both first name and last name empty",
			&Updates{
				"",
				"",
			},
			true,
		},
	}

	for _, c := range cases {
		usr := &User{}

		err := usr.ApplyUpdates(c.updates)
		if (!c.expectError && err != nil) || (c.expectError && err == nil) {
			t.Errorf("\ncase: %s\nexpect error: %v\nerror: %s", c.name, c.expectError, err)
		}
	}
}
