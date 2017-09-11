// This file contains dependency imports
//  which are only required for `go test`.
// Godeps recently introduced a breaking change, which now
//  completely ignors every file ending with `_test.go`,
//  and so the dependencies which are required only for `go test`.
// So, we'll declare those here.
package main

import _ "github.com/stretchr/testify/require"
