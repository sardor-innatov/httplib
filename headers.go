package httplib

import (
	"net/textproto"
	"strings"
)

type Headers map[string][]string

func NewHeaders() Headers {
	return make(Headers)
}

// This func is used for formating Header keys into CamelCase
func (h Headers) formatKey(key string) string {
	return textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(key))
}

func (h Headers) Set(key, value string) {
	k := h.formatKey(key)
	h[k] = []string{strings.TrimSpace(value)}
}

func (h Headers) Add(key, value string) {
	k := h.formatKey(key)
	h[k] = append(h[k], strings.TrimSpace(value))
}

func (h Headers) Get(key string) string {
	k := h.formatKey(key)
	values, exists := h[k]
	if !exists || len(values) == 0 {
		return ""
	}
	return values[0]
}

func (h Headers) Values(key string) []string {
	k := h.formatKey(key)
	return h[k]
}
