package gostree

import (
	"testing"

	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
)

func init() { configureTestLogger() }

func TestSTreeMod(t *testing.T) {

	defer log.Flush()

	Convey("Test clone\n", t, func() {

		s, err := NewSTreeJson(strings.NewReader(`{"key1": "val1", "key.2": 1234, "key3": {"key4": true, "key5": -12.34}}`))
		So(err, ShouldBeNil)

		c, err := s.clone()
		So(err, ShouldBeNil)
		s["key1"] = "valMod"

		s3, err := s.STreeVal(".key3")
		s3["key4"] = false

		log.Debugf("Test clone - s: %v", s)
		log.Debugf("Test clone - c: %v", c)

		v1, err := c.StrVal(".key1")
		So(err, ShouldBeNil)
		So(v1, ShouldEqual, "val1")

		v2, err := c.BoolVal(".key3.key4")
		So(err, ShouldBeNil)
		So(v2, ShouldBeTrue)
	})

	Convey("Test SetVal", t, func() {

		json := `
		{
			"key1": "val1",
			"key2": 1234,
			"key3": {
				"key4": true,
				"key5": -12.34,
				"key6": [
					{"key7": "val7", "key8": 88},
					"sliceVal6",
					{"key9": [99, 999.99, 9999]}
				]
			}
		}`

		s, err := NewSTreeJson(strings.NewReader(json))
		So(err, ShouldBeNil)

		s1, err := s.SetVal(".key1", "val1new")
		So(err, ShouldBeNil)
		So(s1.StrValMust(".key1"), ShouldEqual, "val1new")

		s2, err := s.SetVal(".key3.key5", "val5new")
		So(err, ShouldBeNil)
		So(s2.StrValMust(".key3.key5"), ShouldEqual, "val5new")

		s3, err := s.SetVal(".key3.key6[1]", "sliceVal6new")
		So(err, ShouldBeNil)
		So(s3.StrValMust(".key3.key6[1]"), ShouldEqual, "sliceVal6new")

		s4, err := s.SetVal(".key3.key6[0].key8", 888)
		So(err, ShouldBeNil)
		So(s4.IntValMust(".key3.key6[0].key8"), ShouldEqual, 888)

		s5, err := s.SetVal(".key3.key6[2].key9[1]", 9.999)
		So(err, ShouldBeNil)
		So(s5.FloatValMust(".key3.key6[2].key9[1]"), ShouldEqual, 9.999)

		s6, err := s.SetVal(".key3.key6", "val6new")
		So(err, ShouldBeNil)
		So(s6.StrValMust(".key3.key6"), ShouldEqual, "val6new")
	})

	Convey("Test SetVal no path error", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{}`))
		So(err, ShouldBeNil)
		_, err = s.setPathVal(FieldPath([]string{}), 8)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "called with no path")
	})

	Convey("Test SetVal invalid subscript", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{"key3": {"key6": ["sliceVal6"]}}`))
		So(err, ShouldBeNil)
		_, err = s.SetVal(".key3.key6[abc]", 8)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "parsePathComponent")
	})

	Convey("Test SetVal invalid slice index", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{"key3": {"key6": ["sliceVal6"]}}`))
		So(err, ShouldBeNil)
		_, err = s.SetVal(".key3.key6[2]", 8)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "invalid slice index 2")
	})

	Convey("Test SetVal slice traverse error", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{"key3": {"key6": ["sliceVal6"]}}`))
		So(err, ShouldBeNil)
		_, err = s.SetVal(".key3.key6[0].key99", 8)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unable to traverse below slice path component")
	})

	Convey("Test SetVal stree traverse error", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{"key3": {"key6": "val6"}}`))
		So(err, ShouldBeNil)
		_, err = s.SetVal(".key3.key6.key99", 8)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unable to traverse below path component")
	})

	Convey("Test SetVal stree traverse error", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{"key3": {"key6": "val6"}}`))
		So(err, ShouldBeNil)
		_, err = s.SetVal(".key2", 8)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "path component not found")
	})

	Convey("Test SetVal invalid path syntax", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{"key3": {"key6": "val6"}}`))
		So(err, ShouldBeNil)
		_, err = s.SetVal("key3", 8)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "ValueOfPath error")
	})
}
