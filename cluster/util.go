package cluster

import (
	"strings"
)

func fileExt(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

func stripString(s string) string {
	if i := strings.IndexByte(s, 0); i != -1 {
		return s[:i]
	} else {
		return s
	}
}

func fixString(s string, fix_length int) string {
	buff := make([]byte, fix_length)
	l := len(s)
	if l > fix_length {
		l = fix_length
	}
	for i := 0; i < l; i++ {
		buff[i] = s[i]
	}
	for i := l; i < fix_length; i++ {
		buff[i] = 0
	}
	return string(buff)
}
