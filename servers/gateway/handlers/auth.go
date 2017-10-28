package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
	"net/http"
	"time"
)

// UsersHandler handles requests for the "users" resource,
// and allows clients to create new user accounts.
func (ctx *HandlerContext) UsersHandler(w http.ResponseWriter, r *http.Request) {
	// Method must be POST.
	if r.Method != "POST" {
		http.Error(w, "expect POST method only", http.StatusMethodNotAllowed)
		return
	}

	// Create an empty User to hold decoded request body.
	newUser := &users.NewUser{}

	err := json.NewDecoder(r.Body).Decode(newUser)
	if err != nil {
		http.Error(w, "error decoding request body: invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate the NewUser.
	err = newUser.Validate()
	if err != nil {
		http.Error(w, fmt.Sprintf("error validating new user: %s", err), http.StatusBadRequest)
		return
	}

	// Ensure there isn't already a user in the user store with the same email address.
	_, err = ctx.UserStore.GetByEmail(newUser.Email)
	if err != nil {
		http.Error(w, "user with the same email already exists", http.StatusBadRequest)
		return
	}

	// Ensure there isn't already a user in the user store with the same user name.
	_, err = ctx.UserStore.GetByUserName(newUser.UserName)
	if err != nil {
		http.Error(w, "user with the same username already exists", http.StatusBadRequest)
		return
	}

	// Insert the new user into the user store.
	user, err := ctx.UserStore.Insert(newUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("error inserting new user: %s", err), http.StatusInternalServerError)
		return
	}

	beginNewSession(ctx, user, w)
}

// UsersMeHandler handles requests for the "current user" resource.
func (ctx *HandlerContext) UsersMeHandler(w http.ResponseWriter, r *http.Request) {
	// Get session state from session store.
	sessionState := &SessionState{}
	sessionID, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting session state: %v", err), http.StatusUnauthorized)
		return
	}

	switch r.Method {

	// Get the current user from the session state and respond with that user encoded as JSON object.
	case "GET":
		w.Header().Add(headerContentType, contentTypeJSON)
		err = json.NewEncoder(w).Encode(sessionState.User)
		if err != nil {
			http.Error(w, "error encoding SessionState Struct to JSON", http.StatusInternalServerError)
			return
		}

	// Update the current user with the JSON in the request body,
	// and respond with the newly updated user, encoded as a JSON object.
	case "PATCH":
		// Get Updates struct from request body.
		updates := &users.Updates{}
		err := json.NewDecoder(r.Body).Decode(updates)
		if err != nil {
			http.Error(w, "error decoding request body: invalid JSON in request body", http.StatusBadRequest)
			return
		}

		// Update in-memory session state.
		sessionState.User.FirstName = updates.FirstName
		sessionState.User.LastName = updates.LastName

		// Update session store.
		err = ctx.SessionStore.Save(sessionID, sessionState)
		if err != nil {
			http.Error(w, fmt.Sprintf("error saving updated session state to session store: %s", err), http.StatusInternalServerError)
			return
		}

		// Update user store.
		err = ctx.UserStore.Update(sessionState.User.ID, updates)
		if err != nil {
			http.Error(w, fmt.Sprintf("error updating user store: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Add(headerContentType, contentTypeJSON)
		err = json.NewEncoder(w).Encode(sessionState.User)
		if err != nil {
			http.Error(w, "error encoding SessionState Struct to JSON", http.StatusInternalServerError)
			return
		}

	// If clients send requests that are neither GET nor PATCH...
	default:
		http.Error(w, "expect GET or PATCH method only", http.StatusMethodNotAllowed)
		return
	}
}

// SessionsHandler handles requests for the "sessions" resource,
// and allows clients to begin a new session using an existing user's credentials.
func (ctx *HandlerContext) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	// Method must be POST.
	if r.Method != "POST" {
		http.Error(w, "expect POST method only", http.StatusMethodNotAllowed)
		return
	}

	// Decode the request body into a users.Credentials struct.
	credentials := &users.Credentials{}
	err := json.NewDecoder(r.Body).Decode(credentials)
	if err != nil {
		http.Error(w, "error decoding request body: invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Get the user with the provided email from the UserStore.
	// If not found, respond with an http.StatusUnauthorized error
	// and the message "invalid credentials".
	user, err := ctx.UserStore.GetByEmail(credentials.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Authenticate the user using the provided password.
	// If that fails, respond with an http.StatusUnauthorized error
	// and the message "invalid credentials".
	err = user.Authenticate(credentials.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	beginNewSession(ctx, user, w)
}

// SessionsMineHandler handles requests for the "current session" resource,
// and allows clients to end that session.
func (ctx *HandlerContext) SessionsMineHandler(w http.ResponseWriter, r *http.Request) {
	// Method must be DELETE.
	if r.Method != "DELETE" {
		http.Error(w, "expect DELETE method only", http.StatusMethodNotAllowed)
		return
	}

	// Get session state from session store.
	sessionState := &SessionState{}
	_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting session state: %v", err), http.StatusUnauthorized)
		return
	}

	// End the current session.
	_, err = sessions.EndSession(r, ctx.SigningKey, ctx.SessionStore)
	if err != nil {
		http.Error(w, fmt.Sprintf("error ending session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("signed out"))
}

// begineNewSession begins a new session
// and respond to the client with the User encoded as a JSON object.
func beginNewSession(ctx *HandlerContext, user *users.User, w http.ResponseWriter) {
	sessionState := SessionState{
		BeginTime: time.Now(),
		User:      user,
	}

	_, err := sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, w)
	if err != nil {
		http.Error(w, fmt.Sprintf("error beginning session: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add(headerContentType, contentTypeJSON)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "error encoding User struct to JSON", http.StatusInternalServerError)
		return
	}
}
