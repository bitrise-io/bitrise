package regexputil

import "regexp"

// NamedFindStringSubmatch ...
func NamedFindStringSubmatch(rexp *regexp.Regexp, text string) (map[string]string, bool) {
	match := rexp.FindStringSubmatch(text)
	if match == nil {
		return nil, false
	}
	result := map[string]string{}
	for i, name := range rexp.SubexpNames() {
		if i != 0 {
			result[name] = match[i]
		}
	}
	return result, true
}
