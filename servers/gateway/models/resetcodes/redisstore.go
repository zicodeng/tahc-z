package resetcodes

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// RedisStore represents a resetcodes.Store backed by Redis.
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

// Save saves the provided email and reset code to the store.
func (rs *RedisStore) Save(email string, resetCode string) error {
	err := rs.Client.Set(email, resetCode, rs.SessionDuration).Err()
	if err != nil {
		return fmt.Errorf("error saving data to Redis: %v", err)
	}

	return nil
}

// Get returns ErrResetCodeNotFound if no reset code is found
// for a given email.
func (rs *RedisStore) Get(email string) error {
	_, err := rs.Client.Get(email).Bytes()
	if err != nil {
		return ErrResetCodeNotFound
	}

	return nil
}

// Delete deletes a reset code associated with the email from the store.
func (rs *RedisStore) Delete(email string) error {
	err := rs.Client.Del(email).Err()
	if err != nil {
		return fmt.Errorf("error deleting data: %v", err)
	}
	return nil
}
