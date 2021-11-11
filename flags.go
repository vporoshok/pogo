package pogo

import "strings"

// Flags is an wrapper of strings slice with handle methods
type Flags []string

// String implements fmt.Stringer
func (flags Flags) String() string {
	return strings.Join(flags, ", ")
}

// Parse text to set of flags
//
// Removes empty and duplicates
func (flags *Flags) Parse(text string) {
	tags := strings.Split(text, ",")
	res := make(Flags, 0, len(tags))
	exists := make(map[string]bool, len(tags))
	for i := range tags {
		tag := strings.TrimSpace(tags[i])
		if tag != "" && !exists[tag] {
			res = append(res, tag)
			exists[tag] = true
		}
	}
	*flags = res
}

// Contain the flag?
func (flags Flags) Contain(flag string) bool {
	for i := range flags {
		if flags[i] == flag {
			return true
		}
	}
	return false
}

// Add the flag
//
// Returns false if flags has already contain the flag.
func (flags *Flags) Add(flag string) bool {
	if flags.Contain(flag) {
		return false
	}
	*flags = append(*flags, flag)
	return true
}

// Remove the flag
//
// Returns false if flags hasn't contain the flag.
func (flags *Flags) Remove(flag string) bool {
	for i := range *flags {
		if (*flags)[i] == flag {
			*flags = append((*flags)[:i], (*flags)[i+1:]...)
			return true
		}
	}
	return false
}
