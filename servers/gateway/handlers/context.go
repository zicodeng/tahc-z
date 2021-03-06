package handlers

import (
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/indexes"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/attempts"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/resetcodes"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
)

// HandlerContext will be a receiver on any of your HTTP
// handler functions that need access to
// globals, such as the key used for signing
// and verifying SessionIDs, the session store
// and the user store.
type HandlerContext struct {
	SigningKey string
	Trie       *indexes.Trie
	// The type is an Store interface
	// rather than an actual Store implementation.
	SessionStore   sessions.Store
	UserStore      users.Store
	AttemptStore   attempts.Store
	ResetCodeStore resetcodes.Store
}

// NewHandlerContext constructs a new HanderContext,
// ensuring that the dependencies are valid values.
func NewHandlerContext(
	signingKey string,
	trie *indexes.Trie,
	sessionStore sessions.Store,
	userStore users.Store,
	attemptStore attempts.Store,
	resetCodeStore resetcodes.Store) *HandlerContext {

	if len(signingKey) == 0 {
		panic("signing key has length of zero")
	}

	if trie == nil {
		panic("no trie found")
	}

	if sessionStore == nil {
		panic("nil session store")
	}

	if userStore == nil {
		panic("nil user store")
	}

	if attemptStore == nil {
		panic("nil attempt store")
	}

	if resetCodeStore == nil {
		panic("nil reset code store")
	}

	return &HandlerContext{signingKey, trie, sessionStore, userStore, attemptStore, resetCodeStore}
}
