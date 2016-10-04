package stree

import (
	"strings"
	"testing"

	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
)

func init() { InitTestLogger() }

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
		So(len(paths), ShouldEqual, 5)
		for i, path := range paths {
			log.Debugf("path[%d] = %s", i, path)
		}
		pathsCheck := []FieldPath{
			ValueOfPathMust(`.key1`),
			ValueOfPathMust(`.key2`),
			ValueOfPathMust(`.key\.3.key4`),
			ValueOfPathMust(`.key\.3.key5`),
			ValueOfPathMust(`.key\.3.key6.key7`),
		}
		var m map[string]bool = make(map[string]bool)
		for _, path := range pathsCheck {
			m[path.String()] = false
		}
		for _, path := range paths {
			m[path.String()] = true
		}
		So(len(m), ShouldEqual, len(pathsCheck))
		for _, v := range m {
			So(v, ShouldBeTrue)
		}
	})
}
