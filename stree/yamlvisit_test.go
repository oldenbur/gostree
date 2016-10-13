package stree

import (
	"testing"

	log "github.com/cihub/seelog"
	T "github.com/oldenbur/sql-parser/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"strings"
)

func init() { T.ConfigureTestLogger() }

type MockArgsProducer map[string]string

func (a MockArgsProducer) String(key string) string {
	if val, ok := a[key]; ok {
		return val
	}
	return ""
}

var yamlData string = `
product:
  sku         : BL394D
  quantity    : 4
  subproduct:
    subsku : abc
tax  : 251.42
total: 4443.52
testlist:
- item1: v1
  item1.2: v1.2
- item2: v2
  item2.2: v2.2
  item2.3:
  - subitem2: sv2
  - subitem2.2: sv2.2
last: thing
`

func TestYamlVisit(t *testing.T) {

	defer log.Flush()

	Convey("Test yamlNav\n", t, func() {

		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		visitor := func(header, val, key string) {
			v, err := s.Val(key)
			if err != nil {
				log.Debugf("s.Val(%s) error: %v", key, err)
			}
			So(err, ShouldBeNil)

			if val != "" {
				if IsFloat(v) {
					f, err := strconv.ParseFloat(val, 64)
					So(err, ShouldBeNil)
					So(v, ShouldEqual, f)
				} else if IsInt(v) {
					i, err := strconv.Atoi(val)
					So(err, ShouldBeNil)
					So(v, ShouldEqual, i)
				} else {
					So(PrintValue(v), ShouldEqual, val)
				}
			}
		}
		err = YamlVisit(strings.NewReader(yamlData), visitor)
		So(err, ShouldBeNil)
	})
}
