package stree

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
}
