package gostree

import (
	"testing"

	log "github.com/cihub/seelog"
	"io/ioutil"
	"strings"

	. "github.com/smartystreets/goconvey/convey"
)

func init() { configureTestLogger() }

func TestStruct(t *testing.T) {

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
intList    :
- 1
- 2
- 3
strList    :
- abc
- def
- ghi
comments: >
  Late afternoon is best.
  Backup contact is Nancy
  Billsmer @ 338-4338.
`

	Convey("findStructElems", t, func() {
		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		t, err := s.GoStruct("TestStruct")
		So(err, ShouldBeNil)

		u, err := ioutil.ReadAll(t)
		So(err, ShouldBeNil)
		log.Debugf("u: %s", string(u))
		v := string(u)

		So(v, ShouldContainSubstring, "type TestStruct struct {\n")
		So(v, ShouldContainSubstring, "  BoolField bool `yaml:\"boolField\"`\n")
		So(v, ShouldContainSubstring, "  IntField int `yaml:\"intField\"`\n")
		So(v, ShouldContainSubstring, "  FloatField float64 `yaml:\"floatField\"`\n")
		So(v, ShouldContainSubstring, "  StrField string `yaml:\"strField\"`\n")
		So(v, ShouldContainSubstring, "  Product []struct {\n")
		So(v, ShouldContainSubstring, "    Sku string `yaml:\"sku\"`\n")
		So(v, ShouldContainSubstring, "    Quantity int `yaml:\"quantity\"`\n")
		So(v, ShouldContainSubstring, "    Description string `yaml:\"description\"`\n")
		So(v, ShouldContainSubstring, "    Price float64 `yaml:\"price\"`\n")
		So(v, ShouldContainSubstring, "  } `yaml:\"product\"`\n")
		So(v, ShouldContainSubstring, "  IntList []int `yaml:\"intList\"`\n")
		So(v, ShouldContainSubstring, "  StrList []string `yaml:\"strList\"`\n")
		So(v, ShouldContainSubstring, "  Comments string `yaml:\"comments\"`\n")
		So(v, ShouldContainSubstring, "}")

	})

}
