package urlutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testElem struct {
	testParts      []string
	expectedJoined string
}

func TestJoin(t *testing.T) {
	for _, currTestElem := range []testElem{
		testElem{[]string{"https://bitrise.io", "something"}, "https://bitrise.io/something"},
		testElem{[]string{"https://bitrise.io", "something/a"}, "https://bitrise.io/something/a"},
		testElem{[]string{"https://bitrise.io", "something/a/b"}, "https://bitrise.io/something/a/b"},
		testElem{[]string{"https://bitrise.io", "something/a/b/"}, "https://bitrise.io/something/a/b/"},

		testElem{[]string{"https://bitrise.io", "/something"}, "https://bitrise.io/something"},
		testElem{[]string{"https://bitrise.io", "/something/a"}, "https://bitrise.io/something/a"},
		testElem{[]string{"https://bitrise.io", "/something/a/b"}, "https://bitrise.io/something/a/b"},
		testElem{[]string{"https://bitrise.io", "/something/a/b/"}, "https://bitrise.io/something/a/b/"},

		testElem{[]string{"https://bitrise.io/", "/something"}, "https://bitrise.io/something"},
		testElem{[]string{"https://bitrise.io/", "/something/a"}, "https://bitrise.io/something/a"},
		testElem{[]string{"https://bitrise.io/", "/something/a/b"}, "https://bitrise.io/something/a/b"},
		testElem{[]string{"https://bitrise.io/", "/something/a/b/"}, "https://bitrise.io/something/a/b/"},

		testElem{[]string{"https://bitrise.io//", "//something"}, "https://bitrise.io/something"},
		testElem{[]string{"https://bitrise.io//", "//something/a"}, "https://bitrise.io/something/a"},
		testElem{[]string{"https://bitrise.io//", "//something/a/b"}, "https://bitrise.io/something/a/b"},
		testElem{[]string{"https://bitrise.io//", "//something/a/b/"}, "https://bitrise.io/something/a/b/"},

		testElem{[]string{"https://bitrise-steplib-collection.s3.amazonaws.com/steps", "activate-ssh-key", "assets", "icon.svg"}, "https://bitrise-steplib-collection.s3.amazonaws.com/steps/activate-ssh-key/assets/icon.svg"},
		testElem{[]string{"https://bitrise-steplib-collection.s3.amazonaws.com/steps/", "activate-ssh-key", "assets", "icon.svg"}, "https://bitrise-steplib-collection.s3.amazonaws.com/steps/activate-ssh-key/assets/icon.svg"},
		testElem{[]string{"https://bitrise-steplib-collection.s3.amazonaws.com/steps/", "/activate-ssh-key", "assets", "icon.svg"}, "https://bitrise-steplib-collection.s3.amazonaws.com/steps/activate-ssh-key/assets/icon.svg"},
		testElem{[]string{"https://bitrise-steplib-collection.s3.amazonaws.com/steps/", "/activate-ssh-key", "/assets", "icon.svg"}, "https://bitrise-steplib-collection.s3.amazonaws.com/steps/activate-ssh-key/assets/icon.svg"},
		testElem{[]string{"https://bitrise-steplib-collection.s3.amazonaws.com/steps/", "/activate-ssh-key", "/assets", "/icon.svg"}, "https://bitrise-steplib-collection.s3.amazonaws.com/steps/activate-ssh-key/assets/icon.svg"},
	} {
		url, err := Join(currTestElem.testParts...)
		require.Equal(t, nil, err)
		require.Equal(t, currTestElem.expectedJoined, url)
	}

	elems := []string{"https://", "bitrise.io"}
	url, err := Join(elems...)
	require.Equal(t, "No Host defined", err.Error())
	require.Equal(t, "", url)

	elems = []string{}
	url, err = Join(elems...)
	require.Equal(t, "No elements defined to Join", err.Error())
	require.Equal(t, "", url)
}
