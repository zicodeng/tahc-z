package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
	"net/http"
	"time"
)

// UsersHandler handles requests for the users resource,
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
	_, err = ctx.UserStore.GetByEmail(newUser.UserName)
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

	// Begin a new session.
	sessionState := SessionState{
		BeginTime: time.Now(),
		User:      user,
	}

	_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, w)
	if err != nil {
		http.Error(w, fmt.Sprintf("error beginning session: %s", err), http.StatusInternalServerError)
		return
	}

	// Respond to the client with an http.StatusCreated status code,
	// and the users.User struct returned from the user store insert method encoded as a JSON object.
	w.Header().Add(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "error encoding User struct to JSON", http.StatusInternalServerError)
		return
	}
}
