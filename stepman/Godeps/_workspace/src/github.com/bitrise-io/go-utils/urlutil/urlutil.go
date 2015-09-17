package urlutil

import (
	"errors"
	"net/url"
	"strings"
)

// Join ...
func Join(elems ...string) (string, error) {
	if len(elems) < 1 {
		return "", errors.New("No elements defined to Join")
	}

	url, err := url.Parse(elems[0])
	if err != nil {
		return "", err
	}
	schemeStr := url.Scheme
	if schemeStr == "" {
		return "", errors.New("No Scheme defined")
	}

	hostStr := url.Host
	if hostStr == "" {
		return "", errors.New("No Host defined")
	}

	pathStr := ""
	if url.Path != "" {
		pathStr = clearPrefix(url.Path)
		pathStr = clearSuffix(pathStr)
	}

	combined := schemeStr + "://" + hostStr
	if pathStr != "" {
		combined = combined + "/" + pathStr
	}

	for i, e := range elems {
		if i > 0 {
			if i < len(elems)-1 {
				pathStr = clearPrefix(e)
				pathStr = clearSuffix(pathStr)

				combined = combined + "/" + pathStr
			} else {
				pathStr = clearPrefix(e)
				combined = combined + "/" + pathStr
			}
		}
	}

	return combined, nil
}

func clearPrefix(s string) string {
	if strings.HasPrefix(s, "/") {
		s = s[1:len(s)]
		s = clearPrefix(s)
	}
	return s
}

func clearSuffix(s string) string {
	if strings.HasSuffix(s, "/") {
		s = s[0 : len(s)-1]
		s = clearSuffix(s)
	}
	return s
}
