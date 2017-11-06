package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/attempts"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/resetcodes"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
	"net/http"
	"net/smtp"
	"time"
)

// UsersHandler handles requests for the "users" resource,
// and allows clients to create new user accounts.
func (ctx *HandlerContext) UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	// Finds the first 20 users that
	// match the value of the q query string parameter,
	// and respond with those user profiles encoded as a JSON array of objects.
	case "GET":
		// Get session state from session store.
		sessionState := &SessionState{}
		_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
		if err != nil {
			http.Error(w, fmt.Sprintf("error getting session state: %v", err), http.StatusUnauthorized)
			return
		}

		results := []*users.User{}

		q := r.URL.Query().Get("q")
		if len(q) == 0 {
			w.Header().Add(headerContentType, contentTypeJSON)
			err = json.NewEncoder(w).Encode(results)
			if err != nil {
				http.Error(w, "error encoding search results to JSON", http.StatusInternalServerError)
				return
			}
		}

		ctx.Trie.Mx.RLock()
		userIDs := ctx.Trie.Search(20, q)
		ctx.Trie.Mx.RUnlock()

		results, err = ctx.UserStore.ConvertToUsers(userIDs)
		if err != nil {
			http.Error(w, "error converting to users", http.StatusInternalServerError)
			return
		}

		w.Header().Add(headerContentType, contentTypeJSON)
		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			http.Error(w, "error encoding search results to JSON", http.StatusInternalServerError)
			return
		}

	// Create a new user.
	case "POST":
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
		if err == nil {
			http.Error(w, "user with the same email already exists", http.StatusBadRequest)
			return
		}

		// Ensure there isn't already a user in the user store with the same user name.
		_, err = ctx.UserStore.GetByUserName(newUser.UserName)
		if err == nil {
			http.Error(w, "user with the same username already exists", http.StatusBadRequest)
			return
		}

		// Insert the new user into the user store.
		user, err := ctx.UserStore.Insert(newUser)
		if err != nil {
			http.Error(w, fmt.Sprintf("error inserting new user: %s", err), http.StatusInternalServerError)
			return
		}

		// Add this new user to our trie.
		ctx.Trie.Mx.Lock()
		ctx.Trie.Insert(user.Email, user.ID)
		ctx.Trie.Insert(user.UserName, user.ID)
		ctx.Trie.Insert(user.FirstName, user.ID)
		ctx.Trie.Insert(user.LastName, user.ID)
		ctx.Trie.Mx.Unlock()

		beginNewSession(ctx, user, w)

	default:
		http.Error(w, "expect GET or POST method only", http.StatusMethodNotAllowed)
		return
	}
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

		// Remove the user old fields from the trie.
		ctx.Trie.Mx.Lock()
		ctx.Trie.Remove(sessionState.User.FirstName, sessionState.User.ID)
		ctx.Trie.Remove(sessionState.User.LastName, sessionState.User.ID)
		ctx.Trie.Mx.Unlock()

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

		// Insert the updated user fields into the trie.
		ctx.Trie.Mx.Lock()
		ctx.Trie.Insert(sessionState.User.FirstName, sessionState.User.ID)
		ctx.Trie.Insert(sessionState.User.LastName, sessionState.User.ID)
		ctx.Trie.Mx.Unlock()

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
		err := blockRepeatedFailedSignIns(ctx, credentials.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Authenticate the user using the provided password.
	// If that fails, respond with an http.StatusUnauthorized error
	// and the message "invalid credentials".
	err = user.Authenticate(credentials.Password)
	if err != nil {
		err := blockRepeatedFailedSignIns(ctx, credentials.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// If the user signs in successfully,
	// delete Attempt data associated with the email.
	err = ctx.AttemptStore.Delete(credentials.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Add(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "error encoding User struct to JSON", http.StatusInternalServerError)
		return
	}
}

// blockRepeatedFailedSignIns locks an account
// for a short period of time after several failed sign-in attempts.
func blockRepeatedFailedSignIns(ctx *HandlerContext, email string) error {
	attempt := &attempts.Attempt{}
	// Get failed attempt from AttemptStore.
	err := ctx.AttemptStore.Get(email, attempt)

	if err == attempts.ErrAttemptNotFound {
		initAttempt := &attempts.Attempt{
			Count:     1,
			IsBlocked: false,
		}
		err := ctx.AttemptStore.Save(email, initAttempt, attempts.DefaultExpireTime)
		if err != nil {
			return fmt.Errorf("error saving data to Redis")
		}
	} else {
		// If there is an existing Attempt stored in Redis,
		// and its current Count is less than max attempt,
		// increase its Count by one.
		// Otherwise block the user to sign in
		// for this particular email until freeze time is over.
		if attempt.Count < attempts.MaxAttempt {
			attempt.Count++
			err := ctx.AttemptStore.Save(email, attempt, attempts.DefaultExpireTime)
			if err != nil {
				return fmt.Errorf("error saving data to Redis")
			}
		} else {
			// If not blocked yet, block it.
			if !attempt.IsBlocked {
				attempt.IsBlocked = true
				err := ctx.AttemptStore.Save(email, attempt, attempts.BlockTime)
				if err != nil {
					return fmt.Errorf("error saving data to Redis")
				}
			}
			// If this email is already blocked for further sign-in,
			// report error.
			return fmt.Errorf("you have already failed sign-in more than 5 times with this email. Please wait for ten minutes or try different email")
		}
	}

	return nil
}

// ResetCodesHandler handles a reset code request
// and sends a reset code to the email contained in the request.
func (ctx *HandlerContext) ResetCodesHandler(w http.ResponseWriter, r *http.Request) {
	// Method must be POST.
	if r.Method != "POST" {
		http.Error(w, "expect POST method only", http.StatusMethodNotAllowed)
		return
	}

	resetCodeRequest := &resetcodes.ResetCodeRequest{}

	err := json.NewDecoder(r.Body).Decode(resetCodeRequest)
	if err != nil {
		http.Error(w, "error decoding request body: invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Check if the reset request actually contains
	// email that has associated user stored in our database.
	_, err = ctx.UserStore.GetByEmail(resetCodeRequest.Email)
	if err != nil {
		http.Error(w, "no user found with this email", http.StatusBadRequest)
		return
	}

	// Check if the Redis store already contains this reset code.
	err = ctx.ResetCodeStore.Get(resetCodeRequest.Email)
	if err != resetcodes.ErrResetCodeNotFound {
		http.Error(w, "reset code already sent. please check your email inbox", http.StatusBadRequest)
		return
	}

	// Generate a reset code.
	// We will just use session ID generator for our reset code.
	code, err := sessions.NewSessionID(ctx.SigningKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("error generating reset code: %s", err), http.StatusInternalServerError)
		return
	}
	resetCode := string(code)

	// Save this reset code to Redis.
	err = ctx.ResetCodeStore.Save(resetCodeRequest.Email, resetCode)
	if err != nil {
		http.Error(w, fmt.Sprintf("error saving reset code: %s", err), http.StatusInternalServerError)
		return
	}

	// Send this rest code to the provided email.
	// Set up authentication information.
	auth := smtp.PlainAuth("", "tahczclient@gmail.com", "54tahczclient", "smtp.gmail.com")

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{resetCodeRequest.Email}
	msg := []byte("To:" + resetCodeRequest.Email + "\r\n" +
		"Subject: Tahc-Z: Password Reset Code\r\n" +
		resetCode)
	err = smtp.SendMail("smtp.gmail.com:587", auth, "tahczclient@gmail.com", to, msg)
	if err != nil {
		http.Error(w, fmt.Sprintf("error sending reset code: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("password reset code sent"))
}

// ResetPasswordHandler resets the user's password.
func (ctx *HandlerContext) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Method must be PUT.
	if r.Method != "PUT" {
		http.Error(w, "expect PUT method only", http.StatusMethodNotAllowed)
		return
	}

	passwordReset := &resetcodes.PasswordReset{}

	err := json.NewDecoder(r.Body).Decode(passwordReset)
	if err != nil {
		http.Error(w, "error decoding request body: invalid JSON in request body", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if len(email) == 0 {
		http.Error(w, "no email found in the requested URL", http.StatusBadRequest)
		return
	}

	// Make sure this reset code is used within the valid duration.
	err = ctx.ResetCodeStore.Get(email)
	if err == resetcodes.ErrResetCodeNotFound {
		http.Error(w, "reset code expired", http.StatusBadRequest)
		return
	}

	// Validate reset code.
	_, err = sessions.ValidateID(passwordReset.ResetCode, ctx.SigningKey)
	if err != nil {
		http.Error(w, "invalid reset code", http.StatusBadRequest)
		return
	}

	// Password and PasswordConf must match.
	if passwordReset.Password != passwordReset.PasswordConf {
		http.Error(w, "password must match password confirmation", http.StatusBadRequest)
		return
	}

	// Password must be at least 6 characters.
	if len(passwordReset.Password) < 6 {
		http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Get the user with the provided email.
	oldUser, err := ctx.UserStore.GetByEmail(email)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving user data: %s", err), http.StatusBadRequest)
		return
	}

	// Delete the old user.
	err = ctx.UserStore.Delete(oldUser.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting user data: %s", err), http.StatusInternalServerError)
		return
	}

	// Create a new user based on the old user.
	newUser := &users.NewUser{
		Email:        oldUser.Email,
		Password:     passwordReset.Password,
		PasswordConf: passwordReset.PasswordConf,
		UserName:     oldUser.UserName,
		FirstName:    oldUser.FirstName,
		LastName:     oldUser.LastName,
	}

	// Validate the NewUser.
	err = newUser.Validate()
	if err != nil {
		http.Error(w, fmt.Sprintf("error validating new user: %s", err), http.StatusBadRequest)
		return
	}

	// Insert the new user into the user store.
	user, err := ctx.UserStore.Insert(newUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("error inserting new user: %s", err), http.StatusInternalServerError)
		return
	}

	err = ctx.ResetCodeStore.Delete(email)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting data: %s", err), http.StatusInternalServerError)
		return
	}

	beginNewSession(ctx, user, w)
}
