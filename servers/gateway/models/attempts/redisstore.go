package attempts

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// RedisStore represents a attempts.Store backed by Redis.
type RedisStore struct {
	// Redis client used to talk to redis server.
	Client *redis.Client
}

// NewRedisStore constructs a new RedisStore.
func NewRedisStore(client *redis.Client) *RedisStore {

	// Initialize and return a new RedisStore struct.
	if client == nil {
		client = redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		})
	}

	return &RedisStore{
		Client: client,
	}
}

// Save saves the provided email and Attempt to the store.
func (rs *RedisStore) Save(email string, attempt *Attempt, expiry time.Duration) error {
	j, err := json.Marshal(attempt)
	if nil != err {
		return fmt.Errorf("error marshalling struct to JSON: %v", err)
	}

	err = rs.Client.Set(email, j, expiry).Err()
	if err != nil {
		return fmt.Errorf("error saving session state to Redis: %v", err)
	}

	return nil
}

// Get populates attempt with the data previously saved
// for the given email.
func (rs *RedisStore) Get(email string, attempt *Attempt) error {
	val, err := rs.Client.Get(email).Bytes()
	if err != nil {
		return ErrAttemptNotFound
	}

	err = json.Unmarshal(val, attempt)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON to struct: %v", err)
	}

	return nil
}

// Delete deletes all Attempt data associated with the email from the store.
func (rs *RedisStore) Delete(email string) error {
	err := rs.Client.Del(email).Err()
	if err != nil {
		return fmt.Errorf("error deleting data: %v", err)
	}
	return nil
}
