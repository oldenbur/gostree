package gostree

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func (s STree) GoStruct(name string) (io.Reader, error) {

	buf := &bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("type %s struct {\n", name))
	buf, err := s.goStruct("  ", buf)
	buf.WriteString("}")

	return buf, err
}

func (s STree) goStruct(indent string, buf *bytes.Buffer) (*bytes.Buffer, error) {

	keys, err := s.KeyStrings()
	if err != nil {
		return buf, fmt.Errorf("GoStruct KeyStrings error: %v", err)
	}

	for _, key := range keys {

		val, err := s.Val(PathString(key))
		if err != nil {
			return buf, fmt.Errorf("GoStruct Val(%s) error: %v")
		}

		if buf, err = s.printVal(buf, indent, key, "", val); err != nil {
			return buf, err
		}

		buf.WriteString("")
	}

	return buf, nil
}

func (s STree) printVal(buf *bytes.Buffer, indent, key, typePre string, val interface{}) (*bytes.Buffer, error) {

	var err error

	if IsPrimitive(val) {

		buf = s.printStructPrimitive(buf, indent, key, typePre, val)

	} else if IsMap(val) {

		if buf, err = s.printStructSTree(buf, indent, key, typePre, val); err != nil {
			return buf, err
		}

	} else if IsSlice(val) {

		if buf, err = s.printStructSlice(buf, indent, key, typePre, val); err != nil {
			return buf, err
		}

	}

	return buf, nil
}

func (s STree) printStructPrimitive(buf *bytes.Buffer, indent, key, typePre string, val interface{}) *bytes.Buffer {

	valType := fmt.Sprintf("%T", val)
	buf.WriteString(fmt.Sprintf("%s%s %s%s `yaml:\"%s\"`\n", indent, capitalize(key), typePre, valType, key))
	return buf
}

func (s STree) printStructSTree(buf *bytes.Buffer, indent, key, typePre string, val interface{}) (*bytes.Buffer, error) {

	var err error
	buf.WriteString(fmt.Sprintf("%s%s %sstruct {\n", indent, capitalize(key), typePre))
	if sval, ok := val.(STree); !ok {
		return buf, fmt.Errorf("goStruct failed to convert val to STree: %v", val)
	} else {
		buf, err = sval.goStruct(fmt.Sprintf("%s  ", indent), buf)
		if err != nil {
			return buf, err
		}
	}
	buf.WriteString(fmt.Sprintf("%s} `yaml:\"%s\"`\n", indent, key))
	return buf, nil
}

func (s STree) printStructSlice(buf *bytes.Buffer, indent, key, typePre string, val interface{}) (*bytes.Buffer, error) {

	var err error

	if aval, ok := val.([]interface{}); !ok {

		return buf, fmt.Errorf("goStruct failed to convert val to STree: %v", val)

	} else if len(aval) > 0 {

		var abuf *bytes.Buffer = &bytes.Buffer{}
		for i, v := range aval {
			abuf = &bytes.Buffer{}
			if abuf, err = s.printVal(abuf, indent, key, "[]", v); err != nil {
				return buf, fmt.Errorf("printVal error on slice key: %s  index: %d  error: %v", key, i, err)
			}
		}

		buf.WriteString(abuf.String())

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
