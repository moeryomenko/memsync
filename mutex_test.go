package memsync

import (
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/memcached"
)

func TestMutex(t *testing.T) {
	container, _ := gnomock.Start(memcached.Preset())
	defer func() {
		gnomock.Stop(container)
	}()

	addr := container.DefaultAddress()

	cache := memcache.New(addr)
	lockerFactory := New(cache)
	instance1 := lockerFactory.NewMutex("testlock", WithExpiry(time.Second), WithTries(1))
	if err := instance1.Lock(); err != nil {
		t.Fatalf("could not acquire mutex: %s", err)
	}
	defer instance1.Unlock()

	instance2 := lockerFactory.NewMutex("testlock", WithExpiry(time.Second), WithTries(1))
	if err := instance2.Lock(); err == nil {
		t.Fatalf("first instance don't acquire strong lock")
	}
}

func TestMutex2(t *testing.T) {
	container, _ := gnomock.Start(memcached.Preset())
	defer func() {
		gnomock.Stop(container)
	}()

	addr := container.DefaultAddress()

	cache := memcache.New(addr)
	orderCh := make(chan int)
	mutexes := newTestMutexes(cache, "testlock", 4)
	for instance, mutex := range mutexes {
		go func(i int, mutex *Mutex) {
			err := mutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer mutex.Unlock()

			assertAcquired(t, cache, mutex)

			orderCh <- i
		}(instance, mutex)
	}
	for range mutexes {
		<-orderCh
	}
}

func assertAcquired(t *testing.T, client *memcache.Client, mutex *Mutex) {
	item, err := client.Get(mutex.key)
	if err != nil {
		t.Fatalf("could not get mutex token: %s", err)
	}
	token := string(item.Value)
	if token != mutex.token {
		t.Fatalf("expected token %s, got %s", mutex.token, token)
	}
}

func newTestMutexes(client *memcache.Client, key string, instances int) []*Mutex {
	mutexes := make([]*Mutex, instances)
	for i := 0; i < instances; i++ {
		mutexes[i] = &Mutex{
			key:       key,
			expiry:    1 * time.Second,
			tries:     32,
			delayFunc: func(tries int) time.Duration { return time.Second },
			genToken:  genToken,
			factor:    0.01,
			cache:     client,
		}
	}
	return mutexes
}
