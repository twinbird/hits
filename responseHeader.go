package main

import (
	"fmt"
	"strings"
)

type responseHeader struct {
	Field string
	Value string
}

func (r responseHeader) String() string {
	return r.Field + ":" + r.Value
}

type responseHeaders []*responseHeader

func (rh *responseHeaders) String() string {
	return fmt.Sprintf("%s", *rh)
}

func (rh *responseHeaders) Set(value string) error {
	f, v, err := parseResponseHeader(value)
	if err != nil {
		return err
	}

	*rh = append(*rh, &responseHeader{f, v})
	return nil
}

func parseResponseHeader(s string) (field string, value string, err error) {
	p := strings.Index(s, ":")
	if p < 0 {
		return "", "", fmt.Errorf("'%s' is invalid response header format", s)
	}
	return s[:p], strings.TrimSpace(s[p+1:]), nil
}
