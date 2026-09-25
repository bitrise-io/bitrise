// Copyright (c) 2026 Bart Venter <72999113+bartventer@users.noreply.github.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package memcache implements an in-memory cache backend.
//
// It is suitable for testing or ephemeral caching needs, but does not persist data
// across process restarts.
package memcache

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/bartventer/httpcache/store"
	"github.com/bartventer/httpcache/store/driver"
)

const Scheme = "memcache"

//nolint:gochecknoinits // We use init to register the driver.
func init() {
	store.Register(Scheme, driver.DriverFunc(func(u *url.URL) (driver.Conn, error) {
		return Open(), nil
	}))
}

type memCache struct {
	mu    sync.RWMutex
	store map[string][]byte
}

// Open creates a new in-memory cache.
//
// This cache is not persistent and will lose all data when the process exits.
func Open() *memCache {
	return &memCache{
		store: make(map[string][]byte),
	}
}

var _ driver.Conn = (*memCache)(nil)

func (c *memCache) Get(key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	if !ok {
		return nil, errors.Join(
			driver.ErrNotExist,
			fmt.Errorf("memcache: key %q does not exist", key),
		)
	}
	// Return a copy to prevent mutation
	cp := make([]byte, len(val))
	copy(cp, val)
	return cp, nil
}

func (c *memCache) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Store a copy to prevent external mutation
	cp := make([]byte, len(value))
	copy(cp, value)
	c.store[key] = cp
	return nil
}

func (c *memCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[key]; !ok {
		return errors.Join(
			driver.ErrNotExist,
			fmt.Errorf("memcache: key %q does not exist", key),
		)
	}
	delete(c.store, key)
	return nil
}
