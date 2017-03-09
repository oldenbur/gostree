package gostree

import (
	"fmt"
	"sort"
)

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
	return &VisitorBuilder{v: &visitorImpl{
		vp:  func(string, interface{}) error { return nil },
		vtb: func(string, STree) error { return nil },
		vte: func(string, STree) error { return nil },
		vsb: func(string, []interface{}) error { return nil },
		vse: func(string, []interface{}) error { return nil },
	}}
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
	vis := &visitation{v, nil}
	return vis.visitSTree(FieldPath([]string{}), s)
}

type KeySorter func([]string) []string

var KeySorterAlpha KeySorter = func(keys []string) []string { sort.StringSlice(keys).Sort(); return keys }

func (s STree) VisitSorted(v Visitor, sortFunc KeySorter) error {
	vis := &visitation{v, sortFunc}
	return vis.visitSTree(FieldPath([]string{}), s)
}

type visitation struct {
	visitor  Visitor
	sortFunc func(keys []string) []string
}

func (v *visitation) visitSTree(parentKey FieldPath, t STree) error {

	var keys []string
	var err error
	if keys, err = t.KeyStrings(); err != nil {
		return fmt.Errorf("visit KeyStrings error: %v", err)
	}

	if v.sortFunc != nil {
		keys = v.sortFunc(keys)
	}

	if err = v.visitor.VisitSTreeBegin(parentKey.String(), t); err != nil {
		return err
	}
	for _, key := range keys {

		var val interface{}
		if val, err = t.Val(PathString(key)); err != nil {
			return fmt.Errorf("visit Val(%s) error: %v", PathString(key), err)
		}

		if err = v.visitVal(parentKey.append(key), val); err != nil {
			return err
		}
	}
	if err = v.visitor.VisitSTreeEnd(parentKey.String(), t); err != nil {
		return err
	}
	return nil
}

func (v *visitation) visitVal(key FieldPath, val interface{}) error {

	if IsPrimitive(val) {

		return v.visitor.VisitPrimitive(key.String(), val)

	} else if IsMap(val) {

		if sval, ok := val.(STree); !ok {
			return fmt.Errorf("visitVal failed to convert val to STree: %v", val)
		} else {
			return v.visitSTree(key, sval)
		}

	} else if IsSlice(val) {

		if sval, ok := val.([]interface{}); !ok {
			return fmt.Errorf("visitVal failed to convert val to []interface{}: %v", val)
		} else {
			return v.visitSlice(key, sval)
		}

	} else {

		return fmt.Errorf("visitVal unexpected val type: %V", val)

	}
}

func (v *visitation) visitSlice(parentKey FieldPath, a []interface{}) error {

	var err error
	if err = v.visitor.VisitSliceBegin(parentKey.String(), a); err != nil {
		return err
	}

	keyBase := parentKey[len(parentKey)-1]
	for i, aval := range a {

		parentKey[len(parentKey)-1] = fmt.Sprintf("%s[%d]", keyBase, i)
		if err = v.visitVal(parentKey, aval); err != nil {
			return err
		}
	}
	if err = v.visitor.VisitSliceEnd(parentKey.String(), a); err != nil {
		return err
	}

	return nil
}
