package sessions

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const headerAuthorization = "Authorization"
const paramAuthorization = "auth"
const schemeBearer = "Bearer "

// ErrNoSessionID is used when no session ID was found in the Authorization header.
var ErrNoSessionID = errors.New("no session ID found in " + headerAuthorization + " header")

// ErrInvalidScheme is used when the authorization scheme is not supported.
var ErrInvalidScheme = errors.New("authorization scheme not supported")

// BeginSession creates a new SessionID, saves the `sessionState` to the store, adds an
// Authorization header to the response with the SessionID, and returns the new SessionID.
func BeginSession(signingKey string, store Store, sessionState interface{}, w http.ResponseWriter) (SessionID, error) {

	// Create a new SessionID.
	sessionID, err := NewSessionID(signingKey)
	if err != nil {
		return InvalidSessionID, fmt.Errorf("error creating a new session ID: %v", err)
	}

	// Save the sessionState to the store.
	err = store.Save(sessionID, sessionState)
	if err != nil {
		return InvalidSessionID, fmt.Errorf("error saving session state: %v", err)
	}

	// Add a header to the ResponseWriter that looks like this:
	// "Authorization: Bearer <sessionID>"
	// where "<sessionID>" is replaced with the newly-created SessionID.
	w.Header().Add(headerAuthorization, schemeBearer+sessionID.String())

	return sessionID, nil
}

// GetSessionID extracts and validates the SessionID from the request headers.
func GetSessionID(r *http.Request, signingKey string) (SessionID, error) {

	// Get the value of the Authorization header.
	val := r.Header.Get(headerAuthorization)

	// If no Authorization header is present, get the "auth" query string parameter.
	if len(val) == 0 {
		val = r.URL.Query().Get("auth")
		if len(val) == 0 {
			return InvalidSessionID, ErrNoSessionID
		}
	}

	// The value of a valid Authorization header should look like this:
	// "Bearer <sessionID>"
	// If Bearer is missing, return InvalidSessionID and ErrInvalidScheme.
	if !strings.HasPrefix(val, schemeBearer) {
		return InvalidSessionID, ErrInvalidScheme
	}

	// Get the sessionID part.
	sessionIDVal := strings.TrimPrefix(val, schemeBearer)

	// Validate sessionID from the request header.
	sessionID, err := ValidateID(sessionIDVal, signingKey)
	if err != nil {
		return InvalidSessionID, fmt.Errorf("error validating session ID received from request: %v", err)
	}

	return sessionID, nil
}

// GetState extracts the SessionID from the request,
// gets the associated state from the provided store into
// the `sessionState` parameter, and returns the SessionID.
func GetState(r *http.Request, signingKey string, store Store, sessionState interface{}) (SessionID, error) {

	// Get the SessionID from the request.
	sessionID, err := GetSessionID(r, signingKey)
	if err != nil {
		return sessionID, fmt.Errorf("error getting session ID: %v", err)
	}

	// Get the data associated with that SessionID from the store.
	err = store.Get(sessionID, sessionState)
	if err != nil {
		return sessionID, ErrStateNotFound
	}

	return sessionID, nil
}

// EndSession extracts the SessionID from the request,
// and deletes the associated data in the provided store, returning
// the extracted SessionID.
func EndSession(r *http.Request, signingKey string, store Store) (SessionID, error) {

	// Get the SessionID from the request.
	sessionID, err := GetSessionID(r, signingKey)
	if err != nil {
		return sessionID, fmt.Errorf("error getting session ID: %v", err)
	}

	// Delete the data associated with it in the store.
	err = store.Delete(sessionID)
	if err != nil {
		return sessionID, fmt.Errorf("error deleting session state: %v", err)
	}

	return sessionID, nil
}
