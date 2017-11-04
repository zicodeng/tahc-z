package resetcodes

import (
	"time"
)

// CodeDuration represents the duration of the reset code
// that will be considered as valid. Once the duration passes,
// the reset code will be invalid and then deleted.
const CodeDuration = time.Minute * 5

// ResetCodeRequest represents a request sent by the user
// to request for a password reset code.
type ResetCodeRequest struct {
	Email string `json:"email"`
}

// PasswordReset represents a password reset request.
type PasswordReset struct {
	ResetCode    string `json:"resetCode"`
	Password     string `json:"password"`
	PasswordConf string `json:"passwordConf"`
}
