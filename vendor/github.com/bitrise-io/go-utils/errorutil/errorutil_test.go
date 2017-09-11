package errorutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsExitStatusErrorStr(t *testing.T) {
	// --- Should match ---
	require.Equal(t, true, IsExitStatusErrorStr("exit status 1"))
	require.Equal(t, true, IsExitStatusErrorStr("exit status 0"))
	require.Equal(t, true, IsExitStatusErrorStr("exit status 2"))
	require.Equal(t, true, IsExitStatusErrorStr("exit status 11"))
	require.Equal(t, true, IsExitStatusErrorStr("exit status 111"))
	require.Equal(t, true, IsExitStatusErrorStr("exit status 999"))

	// --- Should not match ---
	require.Equal(t, false, IsExitStatusErrorStr("xit status 1"))
	require.Equal(t, false, IsExitStatusErrorStr("status 1"))
	require.Equal(t, false, IsExitStatusErrorStr("exit status "))
	require.Equal(t, false, IsExitStatusErrorStr("exit status"))
	require.Equal(t, false, IsExitStatusErrorStr("exit status 2112"))
	require.Equal(t, false, IsExitStatusErrorStr("exit status 21121"))

	// prefixed
	require.Equal(t, false, IsExitStatusErrorStr(".exit status 1"))
	require.Equal(t, false, IsExitStatusErrorStr(" exit status 1"))
	require.Equal(t, false, IsExitStatusErrorStr("error: exit status 1"))
	// postfixed
	require.Equal(t, false, IsExitStatusErrorStr("exit status 1."))
	require.Equal(t, false, IsExitStatusErrorStr("exit status 1 "))
	require.Equal(t, false, IsExitStatusErrorStr("exit status 1 - something else"))
	require.Equal(t, false, IsExitStatusErrorStr("exit status 1 2"))

	// other
	require.Equal(t, false, IsExitStatusErrorStr("-exit status 211-"))
	require.Equal(t, false, IsExitStatusErrorStr("something else: exit status 1"))
}
