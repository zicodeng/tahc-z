package handlers

import (
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"time"
)

// SessionState represents session state for an authenticated user.
type SessionState struct {
	// Time struct should be passed as value not pointer.
	BeginTime time.Time
	User      *users.User
}
