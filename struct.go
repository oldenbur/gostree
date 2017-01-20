package gostree

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

var indexRegexp *regexp.Regexp = regexp.MustCompile(`\[\d+\]$`)

func (s STree) GoStruct(structName string) (io.Reader, error) {

	var err error
	var buf *bytes.Buffer = &bytes.Buffer{}

	err = s.Visit(NewVisitorBuilder().
		WithPrimitiveVisitor(func(key string, val interface{}) error {

			var name string = ValueOfPathMust(key).last()

			var skip bool
			var typePre string
			if skip, name, typePre = s.sliceSetup(key, name); skip {
				return nil
			}

			valType := fmt.Sprintf("%T", val)
			buf.WriteString(fmt.Sprintf("%s%s %s%s `yaml:\"%s\"`\n", indent(key), capitalize(name), typePre, valType, name))
			return nil
		}).
		WithSTreeBeginVisitor(func(key string, val STree) error {

			var skip bool
			var typePre string
			var name string = structName
			if len(key) > 0 {
				name = ValueOfPathMust(key).last()
				if skip, name, typePre = s.sliceSetup(key, name); skip {
					return nil
				}
				buf.WriteString(fmt.Sprintf("%s%s %sstruct {\n", indent(key), capitalize(name), typePre))
			} else {
				buf.WriteString(fmt.Sprintf("type %s struct {\n", capitalize(name)))
			}
			return nil
		}).
		WithSTreeEndVisitor(func(key string, val STree) error {

			var skip bool
			var name string = structName
			if len(key) > 0 {
				name = ValueOfPathMust(key).last()
				if skip, name, _ = s.sliceSetup(key, name); skip {
					return nil
				}
				buf.WriteString(fmt.Sprintf("%s} `yaml:\"%s\"`\n", indent(key), name))
			} else {
				buf.WriteString(fmt.Sprintf("}\n"))
			}
			return nil
		}).
		Visitor(),
	)

	return buf, err
}

func (s STree) sliceSetup(key, name string) (bool, string, string) {

	typePre := ""
	if indexRegexp.MatchString(name) {
		typePre = "[]"
	}

	return !s.isFirst(key), indexRegexp.ReplaceAllString(name, ""), typePre
}

func (s STree) isFirst(key string) bool {
	p := ValueOfPathMust(key)
	if len(p) < 1 {
		return true
	} else if _, idx, err := s.parsePathComponent(p[0]); err != nil {
		return true
	} else if idx == 0 || (idx < 0 && len(p) > 1) {
		return s.isFirst(p.shift().String())
	} else {
		return (idx < 0)
	}
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

const singleIndent string = "  "

func indent(key string) string {
	result := ""
	for range ValueOfPathMust(key) {
		result = result + singleIndent
	}
	return result
}
