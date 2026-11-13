package indexgen

// GenerateFromSteplibCloneForTest exposes the unexported fs.FS generator to
// test code only (it is compiled into indexgen's test binary, not the
// production package). The reader integration test uses it to produce a real
// inventory to serve; the only production generation entry point stays Generate.
var GenerateFromSteplibCloneForTest = generateFromSteplibClone
