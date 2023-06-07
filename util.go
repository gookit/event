package event

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// MatchNodePath check for a string.
//
// From the gookit/goutil/strutil.MatchNodePath()
//
// Use on pattern:
//   - `*` match any to sep
//   - `**` match any to end. only allow at start or end on pattern.
func matchNodePath(pattern, s string, sep string) bool {
	if pattern == Wildcard {
		return true
	}

	if i := strings.Index(pattern, AllNode); i >= 0 {
		if i == 0 { // at start
			return strings.HasSuffix(s, pattern[2:])
		}
		return strings.HasPrefix(s, pattern[:len(pattern)-2])
	}

	// eg: "eve.some.*.*" -> match "eve.some.thing.run" "eve.some.thing.do"
	pattern = strings.Replace(pattern, sep, "/", -1)
	s = strings.Replace(s, sep, "/", -1)
	ok, err := path.Match(pattern, s)
	if err != nil {
		ok = false
	}
	return ok
}

// regex for check good event name.
var goodNameReg = regexp.MustCompile(`^[a-zA-Z][\w-.*]*$`)

// goodName check event name is valid.
func goodName(name string, isReg bool) string {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("event: the event name cannot be empty")
	}

	// on add listener
	if isReg {
		if name == AllNode || name == Wildcard {
			return Wildcard
		}
		if strings.HasPrefix(name, AllNode) {
			return name
		}
	}

	if !goodNameReg.MatchString(name) {
		panic(`event: name is invalid, must match regex:` + goodNameReg.String())
	}
	return name
}

func panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}
