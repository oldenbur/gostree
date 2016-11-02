package stree

import (
	"sort"
	"strings"
	"testing"

	log "github.com/cihub/seelog"
	T "github.com/oldenbur/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

func init() { T.ConfigureTestLogger() }

func TestSTreeFieldPaths(t *testing.T) {

	Convey("String", t, func() {
		So(ValueOfPathMust(`.key1`).String(), ShouldEqual, `.key1`)
		So(ValueOfPathMust(`.key\.1`).String(), ShouldEqual, `.key\.1`)
		So(ValueOfPathMust(`.key1.key2.key3`).String(), ShouldEqual, `.key1.key2.key3`)
		So(ValueOfPathMust(`.key\.1.key\.2.key3`).String(), ShouldEqual, `.key\.1.key\.2.key3`)
	})

	Convey("ValueOfPath", t, func() {

		p, err := ValueOfPath(`.key1`)
		So(err, ShouldBeNil)
		So(p, ShouldResemble, FieldPath{"key1"})

		p, err = ValueOfPath(`.key1.key\.2`)
		So(err, ShouldBeNil)
		So(p, ShouldResemble, FieldPath{"key1", "key.2"})

		p, err = ValueOfPath(`.key\.1.key\.2.key3\.`)
		So(err, ShouldBeNil)
		So(p, ShouldResemble, FieldPath{"key.1", "key.2", "key3."})

		p, err = ValueOfPath(`key1.key2`)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "lacks prefix")
	})

	Convey("FieldPaths", t, func() {

		json := `{
		"key1": "val1",
		"key2": 1234,
		"key.3": {
			"key4": true,
			"key5": -12.34,
			"key6": {
				"key7": [1, 2, 3]
			}
		}}`

		s, err := NewSTreeJson(strings.NewReader(json))
		So(err, ShouldBeNil)

		paths := s.FieldPaths()
		So(len(paths), ShouldEqual, 7)
		verifyPaths(paths, []FieldPath{
			ValueOfPathMust(`.key1`),
			ValueOfPathMust(`.key2`),
			ValueOfPathMust(`.key\.3.key4`),
			ValueOfPathMust(`.key\.3.key5`),
			ValueOfPathMust(`.key\.3.key6.key7[0]`),
			ValueOfPathMust(`.key\.3.key6.key7[1]`),
			ValueOfPathMust(`.key\.3.key6.key7[2]`),
		})
	})

	Convey("FieldPaths with slice", t, func() {

		json := `{
		"key1": "val1",
		"key2": [
			1234,
			"abc",
			{
				"key4": -12.34,
				"key3": true,
				"key5": [1, 2, 3]
			}
		]
		}`

		s, err := NewSTreeJson(strings.NewReader(json))
		So(err, ShouldBeNil)

		paths := s.FieldPaths()
		verifyPaths(paths, []FieldPath{
			ValueOfPathMust(`.key1`),
			ValueOfPathMust(`.key2[0]`),
			ValueOfPathMust(`.key2[1]`),
			ValueOfPathMust(`.key2[2].key3`),
			ValueOfPathMust(`.key2[2].key4`),
			ValueOfPathMust(`.key2[2].key5[0]`),
			ValueOfPathMust(`.key2[2].key5[1]`),
			ValueOfPathMust(`.key2[2].key5[2]`),
		})
	})

	Convey("FieldPaths with nested slice", t, func() {

		json := `{
			"key1": [
				[1, [2, 3, 4], 3],
				4,
				{
					"key2": -12.34,
					"key3": ["a", ["b"], "c"]
				}
			]
		}`

		s, err := NewSTreeJson(strings.NewReader(json))
		So(err, ShouldBeNil)

		paths := s.FieldPaths()
		verifyPaths(paths, []FieldPath{
			ValueOfPathMust(`.key1[0][0]`),
			ValueOfPathMust(`.key1[0][1][0]`),
			ValueOfPathMust(`.key1[0][1][1]`),
			ValueOfPathMust(`.key1[0][1][2]`),
			ValueOfPathMust(`.key1[0][2]`),
			ValueOfPathMust(`.key1[1]`),
			ValueOfPathMust(`.key1[2].key2`),
			ValueOfPathMust(`.key1[2].key3[0]`),
			ValueOfPathMust(`.key1[2].key3[1][0]`),
			ValueOfPathMust(`.key1[2].key3[2]`),
		})
	})
}

func verifyPaths(paths, pathsCheck []FieldPath) {
	if len(paths) != len(pathsCheck) {
		printSorted("paths", paths)
		printSorted("pathsCheck", pathsCheck)
	}
	So(len(paths), ShouldEqual, len(pathsCheck))
	var m map[string]bool = make(map[string]bool)
	for _, path := range pathsCheck {
		m[path.String()] = false
	}
	for _, path := range paths {
		m[path.String()] = true
	}
	So(len(m), ShouldEqual, len(pathsCheck))
	for _, v := range m {
		if !v {
			printSorted("paths", paths)
			printSorted("pathsCheck", pathsCheck)
		}
		So(v, ShouldBeTrue)
	}
}

func printSorted(msg string, paths []FieldPath) {
	log.Debug("paths")
	sort.Sort(FieldPaths(paths))
	for _, p := range paths {
		log.Debug("  ", p)
	}
}

type FieldPaths []FieldPath

func (p FieldPaths) Len() int           { return len(p) }
func (p FieldPaths) Less(i, j int) bool { return p[i].String() < p[j].String() }
func (p FieldPaths) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
