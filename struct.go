package gostree

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func (s STree) GoStruct(indent string) (io.Reader, error) {

	buf := &bytes.Buffer{}

	keys, err := s.KeyStrings()
	if err != nil {
		return buf, fmt.Errorf("GoStruct KeyStrings error: %v", err)
	}

	for _, key := range keys {

		val, err := s.Val(PathString(key))
		if err != nil {
			return buf, fmt.Errorf("GoStruct Val(%s) error: %v")
		}

		if IsBool(val) {
			buf.WriteString(fmt.Sprintf("%s%s bool `yaml:\"%s\"`\n", indent, capitalize(key), key))
		}

		buf.WriteString("")
	}

	return buf, nil
}

func capitalize(s string) string {

	firstChar := true
	return strings.Map(func(r rune) rune {
		if firstChar {
			firstChar = false
			return unicode.ToUpper(r)
		} else {
			return r
		}
	}, s)
}
