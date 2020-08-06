package main

import (
	"strings"
)

// PathSeparator describes a path separating character
const PathSeparator = "/"

// Path describes a path structure
type Path struct {
	Path string
	ID   string
}

// NewPath parses the specified path.
// Returns a new instance of the Path type.
func NewPath(p string) *Path {
	var id string
	// remove leading and trailing separators
	p = strings.Trim(p, PathSeparator)
	// remove separators between path and split into substrings
	s := strings.Split(p, PathSeparator)
	if len(s) > 1 {
		id = s[len(s)-1]
		p = strings.Join(s[:len(s)-1], PathSeparator)
	}
	return &Path{Path: p, ID: id}
}

// HasID checks if the path has an ID
func (p *Path) HasID() bool{
	return len(p.ID) > 0
}