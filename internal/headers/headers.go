package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

const clrf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	n = 0
	done = false
	err = nil
	before, _, ok := bytes.Cut(data, []byte(clrf))
	if !ok {
		return
	}
	if len(before) == 0 { // final clrf
		n = 2
		done = true
		return
	}

	headerStr := string(before)
	keyValue := strings.SplitN(headerStr, ":", 2)
	if len(keyValue) != 2 || keyValue[0] != strings.TrimRight(keyValue[0], " ") {
		err = fmt.Errorf("invalid header: %s", headerStr)
		return
	}
	key := strings.TrimSpace(keyValue[0])
	if err = isValidKey(key); err != nil {
		return
	}
	key = strings.ToLower(key)
	value := strings.TrimSpace(keyValue[1])
	n = len(before) + 2
	h.Set(key, value)
	return
}

func isValidKey(key string) error {
	for _, k := range key {
		special := []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}
		if !(k >= '0' && k <= '9') && !(k >= 'a' && k <= 'z') && !(k >= 'A' && k <= 'Z') && !slices.Contains(special, k) {
			return fmt.Errorf("invalid symbol: %c", k)
		}
	}
	return nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = v + ", " + value
	}
	h[key] = value
}

func (h Headers) Unset(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
	v, ok := h[strings.ToLower(key)]
	return v, ok
}
