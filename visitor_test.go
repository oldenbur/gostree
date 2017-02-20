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

	Convey("test VisitSorted", t, func() {

		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		visited := []string{}

		err = s.VisitSorted(
			NewVisitorBuilder().
				WithPrimitiveVisitor(func(key string, val interface{}) error {
					log.Debugf("primitive - key: %s", key)
					visited = append(visited, key)
					return nil
				}).
				WithSTreeBeginVisitor(func(key string, val STree) error {
					log.Debugf("stree begin - key: %s", key)
					visited = append(visited, key)
					return nil
				}).
				WithSliceBeginVisitor(func(key string, val []interface{}) error {
					log.Debugf("slice being - key: %s", key)
					visited = append(visited, key)
					return nil
				}).
				Visitor(),
			KeySorterAlpha,
		)
		So(err, ShouldBeNil)
		So(visited, ShouldResemble, []string{
			".boolField",
			".floatField",
			".intField",
			".product",
			".product[0]",
			".product[0].description",
			".product[0].price",
			".product[0].quantity",
			".product[0].sku",
			".product[1]",
			".product[1].description",
			".product[1].price",
			".product[1].quantity",
			".product[1].sku",
			".strField",
		})
	})

	Convey("test Visitor doing nothing", t, func() {

		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		err = s.Visit(NewVisitorBuilder().Visitor())
		So(err, ShouldBeNil)
	})

	Convey("test Visitor Ends", t, func() {

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

}
