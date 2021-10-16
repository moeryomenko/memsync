package memsync

import (
	"bytes"
	"encoding/base32"
	"errors"
	"math/rand"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

var ErrFailed = errors.New("memsync: failed to acquire lock")

// DelayFunc is used to decide the amount of time to wait between retries.
type DelayFunc func(tries int) time.Duration

// A Mutex is a distributed mutual exclusion lock.
type Mutex struct {
	key    string
	expiry time.Duration

	tries     int
	delayFunc DelayFunc

	factor float64

	genToken func() (string, error)
	token    string
	until    time.Time

	cache *memcache.Client
}

// Key returns mutex key.
func (m *Mutex) Key() string {
	return m.key
}

// Token returns the current random token. The token will be empty until a lock is acquired.
func (m *Mutex) Token() string {
	return m.token
}

// Until returns the time of validity of acquired lock. The token will be zero token until a lock is acquired.
func (m *Mutex) Until() time.Time {
	return m.until
}

// Lock tries lock mutex, in case it returns an error on failure.
func (m *Mutex) Lock() error {
	token, err := m.genToken()
	if err != nil {
		return err
	}

	for i := 0; i < m.tries; i++ {
		if i != 0 {
			time.Sleep(m.delayFunc(i))
		}

		start := time.Now()

		err := m.acquire(token)
		if err != nil {
			continue
		}

		now := time.Now()
		until := now.Add(m.expiry - now.Sub(start) - time.Duration(int64(float64(m.expiry)*m.factor)))
		if now.Before(until) {
			m.token = token
			m.until = until
			return nil
		}

		_ = m.release(token)
	}
	return ErrFailed
}

func (m *Mutex) Unlock() error {
	return m.release(m.token)
}

func (m *Mutex) acquire(token string) error {
	return m.cache.Add(&memcache.Item{
		Key:        m.key,
		Value:      []byte(token),
		Expiration: int32(m.expiry / time.Second),
	})
}

func (m *Mutex) release(token string) error {
	value, err := m.cache.Get(m.key)
	if err != nil {
		return err
	}
	if bytes.Compare(value.Value, []byte(token)) == 0 {
		return m.cache.Delete(m.key)
	}
	return nil
}

func genToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(b), nil
}
