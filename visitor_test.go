package gostree

import (
	"fmt"
	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func init() { configureTestLogger() }

func TestVisitor(t *testing.T) {

	defer log.Flush()

	var yamlData string = `
---
boolField  : false
intField   : 432
floatField : 8765.43
strField   : Super Hoop
product:
- sku         : BL394D
  quantity    : 4
  description : Basketball
  price       : 450.00
- sku         : BL4438H
  quantity    : 1
  description : Super Hoop
  price       : 2392.00
`

	Convey("test Visitor Begins", t, func() {

		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		paths := s.FieldPaths()
		visited := map[string]string{}
		for _, p := range paths {
			visited[p.String()] = ""
		}

		err = s.Visit(NewVisitorBuilder().
			WithPrimitiveVisitor(func(key string, val interface{}) error {
				visited[key] = key
				fmt.Printf("primitive - %s: %v", key, val)
				return nil
			}).
			WithSTreeBeginVisitor(func(key string, val STree) error {
				visited[key] = key
				return nil
			}).
			WithSliceBeginVisitor(func(key string, val []interface{}) error {
				visited[key] = key
				return nil
			}).
			Visitor(),
		)
		So(err, ShouldBeNil)

		for _, p := range paths {
			So(visited[p.String()], ShouldEqual, p.String())
		}
	})

	Convey("test Visitor Begins", t, func() {

		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		paths := s.FieldPaths()
		visited := map[string]string{}
		for _, p := range paths {
			visited[p.String()] = ""
		}

		err = s.Visit(NewVisitorBuilder().
			WithPrimitiveVisitor(func(key string, val interface{}) error {
				visited[key] = key
				return nil
			}).
			WithSTreeEndVisitor(func(key string, val STree) error {
				visited[key] = key
				return nil
			}).
			WithSliceEndVisitor(func(key string, val []interface{}) error {
				visited[key] = key
				return nil
			}).
			Visitor(),
		)
		So(err, ShouldBeNil)

		for _, p := range paths {
			So(visited[p.String()], ShouldEqual, p.String())
		}
	})

	Convey("test Visitor doing nothing", t, func() {

		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		err = s.Visit(NewVisitorBuilder().Visitor())
		So(err, ShouldBeNil)
	})

}
