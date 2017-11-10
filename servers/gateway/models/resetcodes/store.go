package resetcodes

import (
	"errors"
)

// ErrResetCodeNotFound is returned if no ResetCode is found
// for a given email.
var ErrResetCodeNotFound = errors.New("reset code not found")

// Store stores a reset code for a given email.
type Store interface {
	Save(email string, resetCode string) error

	Get(email string) error

	Delete(email string) error
}
