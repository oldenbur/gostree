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

func (s STree) GoStruct2(name string) (io.Reader, error) {

	var err error
	var buf *bytes.Buffer = &bytes.Buffer{}

	v := NewVisitorBuilder().
		WithPrimitiveVisitor(func(key string, val interface{}) error {

			return nil
		}).
		WithSTreeBeginVisitor(func(key string, val STree) error {

			return nil
		}).
		WithSTreeEndVisitor(func(key string, val STree) error {

			return nil
		}).
		WithSliceBeginVisitor(func(key string, val []interface{}) error {

			return nil
		}).
		WithSliceEndVisitor(func(key string, val []interface{}) error {

			return nil
		}).
		Visitor()

	err = s.Visit(v)

	return buf, err
}

type Visitor interface {
	VisitPrimitive(key string, val interface{}) error
	VisitSTreeBegin(key string, val STree) error
	VisitSTreeEnd(key string, val STree) error
	VisitSliceBegin(key string, val []interface{}) error
	VisitSliceEnd(key string, val []interface{}) error
}

type visitorImpl struct {
	vp  func(string, interface{}) error
	vtb func(string, STree) error
	vte func(string, STree) error
	vsb func(string, []interface{}) error
	vse func(string, []interface{}) error
}

func (v *visitorImpl) VisitPrimitive(key string, val interface{}) error {
	return v.vp(key, val)
}
func (v *visitorImpl) VisitSTreeBegin(key string, val STree) error {
	return v.vtb(key, val)
}
func (v *visitorImpl) VisitSTreeEnd(key string, val STree) error {
	return v.vte(key, val)
}
func (v *visitorImpl) VisitSliceBegin(key string, val []interface{}) error {
	return v.vsb(key, val)
}
func (v *visitorImpl) VisitSliceEnd(key string, val []interface{}) error {
	return v.vse(key, val)
}

type VisitorBuilder struct {
	v *visitorImpl
}

func NewVisitorBuilder() *VisitorBuilder {
	return &VisitorBuilder{}
}
func (b *VisitorBuilder) WithPrimitiveVisitor(f func(string, interface{}) error) *VisitorBuilder {
	b.v.vp = f
	return b
}
func (b *VisitorBuilder) WithSTreeBeginVisitor(f func(string, STree) error) *VisitorBuilder {
	b.v.vtb = f
	return b
}
func (b *VisitorBuilder) WithSTreeEndVisitor(f func(string, STree) error) *VisitorBuilder {
	b.v.vte = f
	return b
}
func (b *VisitorBuilder) WithSliceBeginVisitor(f func(string, []interface{}) error) *VisitorBuilder {
	b.v.vsb = f
	return b
}
func (b *VisitorBuilder) WithSliceEndVisitor(f func(string, []interface{}) error) *VisitorBuilder {
	b.v.vse = f
	return b
}
func (b *VisitorBuilder) Visitor() Visitor {
	return b.v
}

func (s STree) Visit(v Visitor) error {

	return s.visitSTree(FieldPath([]string{}), v, s)
}

func (s STree) visitSTree(parentKey FieldPath, v Visitor, t STree) error {

	var keys []string
	var err error
	if keys, err = t.KeyStrings(); err != nil {
		return fmt.Errorf("visit KeyStrings error: %v", err)
	}

	if err = v.VisitSTreeBegin(parentKey.String(), t); err != nil {
		return err
	}
	for _, key := range keys {

		var val interface{}
		if val, err = s.Val(PathString(key)); err != nil {
			return fmt.Errorf("visit Val(%s) error: %v")
		}

		if err = s.visitVal(parentKey.append(key), v, val); err != nil {
			return err
		}
	}
	if err = v.VisitSTreeEnd(parentKey.String(), t); err != nil {
		return err
	}

	return nil
}

func (s STree) visitVal(key FieldPath, v Visitor, val interface{}) error {

	if IsPrimitive(val) {

		return v.VisitPrimitive(key.String(), val)

	} else if IsMap(val) {

		if sval, ok := val.(STree); !ok {
			return fmt.Errorf("visitVal failed to convert val to STree: %v", val)
		} else {
			return s.visitSTree(key, v, sval)
		}

	} else if IsSlice(val) {

		if sval, ok := val.([]interface{}); !ok {
			return fmt.Errorf("visitVal failed to convert val to []interface{}: %v", val)
		} else {
			return s.visitSlice(key, v, sval)
		}

	} else {

		return fmt.Errorf("visitVal unexpected val type: %V", val)

	}
}

func (s STree) visitSlice(parentKey FieldPath, v Visitor, a []interface{}) error {

	var err error
	if err = v.VisitSliceBegin(parentKey.String(), a); err != nil {
		return err
	}
	for i, aval := range a {

		parentKey[len(parentKey)-1] = fmt.Sprintf("%s[%d]", parentKey[len(parentKey)-1], i)
		if err = s.visitVal(parentKey, v, aval); err != nil {
			return err
		}
	}
	if err = v.VisitSliceEnd(parentKey.String(), a); err != nil {
		return err
	}

	return nil
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
