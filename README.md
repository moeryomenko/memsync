# Memsync [![Go Reference](https://pkg.go.dev/badge/github.com/moeryomenko/memsync.svg)](https://pkg.go.dev/github.com/moeryomenko/memsync)

Memsync provides a Memcached-based distributed mutual exclusion lock implementation for Go.

## Installation

Install Memsync using:

	$ go get github.com/moeryomenko/memsync

## Usage

```go
package main

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/moeryomenko/memsync"
)

func main() {
	// Create a memcached client.
	client := memcache.New("localhost:112211")

	// Create an instance of memsync to be used to obtain a mutual exclusion lock.
	ms := memsync.New(client)

	// Obtaion a new mutex by using the same for all instances wanting the same lock.
	mx := rw.NewMutex(
		"examplelock",           // exlusive key for lock.
		WithExpiry(time.Second), // sets custom expiry of mutex.
		WithTries(8),            // sets numbers of times lock acquire is attempted.
	)

	// Obtain a lock for given mutex. After this is successful, no one else
	// can obtain the same lock (the same mutex key) until we unlock it.
	for err := mutex.Lock(); err != nil; {}

	// Do work that requires the lock.

	// Release the lock so other instances can obtain a lock.
	if err := mutex.Unlock(); err != nil {
		panic("unlock failed")
	}
}
```

## License

Memsync is primarily distributed under the terms of both the MIT license and the Apache License (Version 2.0).

See [LICENSE-APACHE](LICENSE-APACHE) and/or [LICENSE-MIT](LICENSE-MIT) for details.
