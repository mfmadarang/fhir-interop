package hl7v2

import "strings"

type Field string

func (f Field) String() string {
	return string(f)
}

func (f Field) Component(i int) string {
	parts := strings.Split(string(f), "^")
	idx := i - 1
	if idx < 0 || idx >= len(parts) {
		return ""
	}
	return parts[idx]
}
