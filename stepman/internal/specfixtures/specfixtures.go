// Package specfixtures provides shared, embedded test fixtures for the V2
// steplib steplibindex/indexgen packages — currently a small sample steplib clone
// exposed as an fs.FS.
//
// It lives under internal/ so only this module can import it (consumers of the
// stepman module never see it), and it serves the fixtures via //go:embed so any
// package's tests can reach them regardless of the test's working directory —
// unlike a per-package testdata/ dir, which is only conveniently reachable from
// its own package.
package specfixtures

import (
	"embed"
	"io/fs"
)

//go:embed testdata/steplib
var files embed.FS

// SteplibClone returns the sample steplib clone (steplib.yml + steps/<id>/...)
// rooted so it can be passed straight to the indexgen generator.
func SteplibClone() fs.FS {
	sub, err := fs.Sub(files, "testdata/steplib")
	if err != nil {
		// The embedded path is a compile-time constant, so Sub cannot fail.
		panic("specfixtures: " + err.Error())
	}
	return sub
}
