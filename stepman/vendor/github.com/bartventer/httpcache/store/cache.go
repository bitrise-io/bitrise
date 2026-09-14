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

// Package store provides a registry and entry point for cache backends.
//
// This package allows you to register cache drivers and open cache connections
// using a DSN string. It acts as a facade over the internal registry and the
// driver interfaces defined in the [driver] subpackage.
//
// Most users won't interact with this package directly, but will instead use it
// indirectly through higher-level caching abstractions.
// For more details on registering drivers and opening connections, see the
// documentation for the [Register] and [Open] functions.
package store

import (
	"github.com/bartventer/httpcache/store/driver"
	"github.com/bartventer/httpcache/store/internal/registry"
)

// Register makes a driver implementation available by the provided name (e.g.,
// "memcache", "fscache", etc.). The name corresponds to the scheme component of
// a DSN string used in [Open] to identify the driver to use.
// If Register is called twice with the same name or if driver is nil, it
// panics.
func Register(name string, driver driver.Driver) {
	registry.Default().RegisterDriver(name, driver)
}

// Open establishes a connection to a registered driver, as specified in the
// provided DSN string.
//
// The DSN (Data Source Name) format follows the pattern:
//
//	scheme://[path]?[query_parameters]
//
// Where:
//   - scheme: The name of the registered driver (e.g., "memcache", "fscache",
//     etc.).
//   - path: An optional path component that can be used by the driver for
//     configuration.
//   - query_parameters: Optional key-value pairs for additional driver-
//     specific settings.
//
// See [Register] for registering drivers.
func Open(dsn string) (driver.Conn, error) {
	return registry.Default().OpenConn(dsn)
}

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	return registry.Default().Drivers()
}
