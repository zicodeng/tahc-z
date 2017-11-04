package sessions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// RedisStore represents a session.Store backed by redis.
type RedisStore struct {
	// Redis client used to talk to redis server.
	Client *redis.Client
	// Used for key expiry time on redis.
	SessionDuration time.Duration
}

// NewRedisStore constructs a new RedisStore.
func NewRedisStore(client *redis.Client, sessionDuration time.Duration) *RedisStore {

	// Initialize and return a new RedisStore struct.
	if client == nil {
		client = redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		})
	}

	return &RedisStore{
		Client:          client,
		SessionDuration: sessionDuration,
	}
}

// Store implementation

// Save saves the provided `sessionState` and associated SessionID to the store.
// The `sessionState` parameter is typically a pointer to a struct containing
// all the data you want to associated with the given SessionID.
func (rs *RedisStore) Save(sid SessionID, sessionState interface{}) error {
	// Marshal the `sessionState` to JSON.
	j, err := json.Marshal(sessionState)
	if nil != err {
		return fmt.Errorf("error marshalling struct to JSON: %v", err)
	}

	// Save it in the redis database,
	// using `sid.getRedisKey()` for the key.
	// Return any errors that occur along the way.
	err = rs.Client.Set(sid.getRedisKey(), j, rs.SessionDuration).Err()
	if err != nil {
		return fmt.Errorf("error saving session state to Redis: %v", err)
	}

	return nil
}

// Get populates `sessionState` with the data previously saved
// for the given SessionID.
func (rs *RedisStore) Get(sid SessionID, sessionState interface{}) error {

	// Use the Pipeline feature of the redis
	// package to do both the get and the reset of the expiry time
	// in just one network round trip.
	pipe := rs.Client.Pipeline()

	// get is a string command that retrieves the previously-saved session state
	// for a given key from redis.
	// It is just a command waiting to be executed.
	get := pipe.Get(sid.getRedisKey())

	// Reset the expiry time, so that it doesn't get deleted until
	// the SessionDuration has elapsed.
	pipe.Expire(sid.getRedisKey(), rs.SessionDuration)

	// Execute all previously queued commands using one client-server roundtrip.
	_, err := pipe.Exec()
	if err != nil {
		return ErrStateNotFound
	}

	// Extract session state data from get command.
	val, err := get.Bytes()

	// Unmarshal it back into the `sessionState` parameter.
	err = json.Unmarshal(val, sessionState)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON to struct: %v", err)
	}

	return nil
}

// Delete deletes all state data associated with the SessionID from the store.
func (rs *RedisStore) Delete(sid SessionID) error {
	// Delete the data stored in redis for the provided SessionID.
	err := rs.Client.Del(sid.getRedisKey()).Err()
	if err != nil {
		return fmt.Errorf("error deleting session state: %v", err)
	}
	return nil
}

// getRedisKey() returns the redis key to use for the SessionID
func (sid SessionID) getRedisKey() string {
	// Convert the SessionID to a string and add the prefix "sid:" to keep
	// SessionID keys separate from other keys that might end up in this
	// redis instance.
	return "sid:" + sid.String()
}
