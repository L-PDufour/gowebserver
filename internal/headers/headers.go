package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func IsValidHeaderKey(s string) bool {
	if len(s) == 0 {
		return false
	}
	const tchar = "!#$%&'*+-.^_`|~"
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune(tchar, r)
	})
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}
	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := strings.ToLower(string(parts[0]))

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	key = strings.ToLower(strings.TrimSpace(key))

	if !IsValidHeaderKey(key) {
		return 0, false, fmt.Errorf("invalid character: %s", key)
	}
	if existing := h[key]; existing != "" {
		h[key] = existing + ", " + string(value)
	} else {
		h[key] = string(value)
	}
	return idx + 2, false, nil
}

func (h Headers) Get(key string) string {
	return h[key]
}
