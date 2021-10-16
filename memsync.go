package memsync

import (
	"math/rand"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

const (
	minRetryDelaySec = 1
	maxRetryDelaySec = 3
)

// Memsync provides a simple method for creating distributed mutexes using memcache.
type Memsync struct {
	cache *memcache.Client
}

// New creates and returns a new Memsync instance.
func New(cache *memcache.Client) *Memsync {
	return &Memsync{cache: cache}
}

// NewMutex returns a new distributed mutex with given key.
func (m *Memsync) NewMutex(key string, options ...Option) *Mutex {
	mu := &Mutex{
		key:    key,
		expiry: 3 * time.Second,
		tries:  4,
		delayFunc: func(tries int) time.Duration {
			return time.Duration(rand.Intn(maxRetryDelaySec-minRetryDelaySec)+minRetryDelaySec) * time.Second
		},
		genToken: genToken,
		factor:   0.01,
		cache:    m.cache,
	}
	for _, opt := range options {
		opt(mu)
	}
	return mu
}

// Option confibures a mutex.
type Option func(*Mutex)

// WithExpiry can be used to set the expiry of mutex to the given token.
func WithExpiry(expiry time.Duration) Option {
	return func(m *Mutex) {
		m.expiry = expiry
	}
}

// WithTries can be used to set the number of times lock acquire is attempted.
func WithTries(tries int) Option {
	return func(m *Mutex) {
		m.tries = tries
	}
}

// WithRetryDelay can be used to set the amount of time to wait between retries.
func WithRetryDelay(delay time.Duration) Option {
	return func(m *Mutex) {
		m.delayFunc = func(_ int) time.Duration {
			return delay
		}
	}
}

// WithRetryDelayFunc can be used to override default delay behavior.
func WithRetryDelayFunc(delayFn DelayFunc) Option {
	return func(m *Mutex) {
		m.delayFunc = delayFn
	}
}

// WithDriftFactor can be used to set the clock drift factor.
func WithDriftFactor(factor float64) Option {
	return func(m *Mutex) {
		m.factor = factor
	}
}

// WithGenTokenFunc can be used to set custom value generator.
func WithGenTokenFunc(genTokenFunc func() (string, error)) Option {
	return func(m *Mutex) {
		m.genToken = genToken
	}
}
