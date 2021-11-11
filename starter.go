package pogo

import (
	"fmt"
	"regexp"
	"strings"
)

// Starter is an pair of border and prefix
type Starter interface {
	Extract(line string) (border, prefix string, ok bool)
}

type plainStarter struct {
	border, prefix string
}

// NewPlainStarter returns new plain starter
func NewPlainStarter(border, prefix string) Starter {
	return plainStarter{border, prefix}
}

// Extract implements Starter.Extract
func (ps plainStarter) Extract(line string) (border, prefix string, ok bool) {
	if strings.HasPrefix(line, ps.border+ps.prefix) {
		return ps.border, ps.prefix, true
	}
	return "", "", false
}

type regexpStarter struct {
	matcher *regexp.Regexp

	border, prefix string
}

// NewRegexpStarter returns new regexp starter
func NewRegexpStarter(border, prefix string) Starter {
	matcher := regexp.MustCompile(fmt.Sprintf(`^(%s)(%s)`, border, prefix))
	return regexpStarter{matcher, border, prefix}
}

// Extract implements Starter.Extract
func (rs regexpStarter) Extract(line string) (border, prefix string, ok bool) {
	if rs.matcher.MatchString(line) {
		submatch := rs.matcher.FindStringSubmatch(line)
		return submatch[1], submatch[2], true
	}
	return "", "", false
}
