package gostree

import (
	"testing"

	log "github.com/cihub/seelog"
	"io/ioutil"
	"strings"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"
)

func init() { configureTestLogger() }

func TestStruct(t *testing.T) {

	defer log.Flush()

	// HACK!!! If this struct changes, the definition of TestStructProto below should change
	// accordingly, along with the test cases.
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
topLevel:
  subStruct:
    subKey1: subVal1
    subKey2: 543
    subList: [6, 5, 4]
    subSubStruct:
      subSubKey1: subSubVal1
`

	Convey("test GoStruct", t, func() {
		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)

		t, err := s.GoStruct("TestStructProto")
		So(err, ShouldBeNil)

		u, err := ioutil.ReadAll(t)
		So(err, ShouldBeNil)
		v := string(u)
		log.Debugf("v: %s", v)

		x := TestStructProto{}
		err = yaml.Unmarshal([]byte(yamlData), &x)
		So(err, ShouldBeNil)
		So(x.BoolField, ShouldBeFalse)

		So(len(x.IntList), ShouldEqual, 3)
		So(x.Product[1].Description, ShouldEqual, "Super Hoop")
		So(x.TopLevel.SubStruct.SubKey1, ShouldEqual, "subVal1")
		So(x.TopLevel.SubStruct.SubList[1], ShouldEqual, 5)
	})

	Convey("test isFirst", t, func() {
		s := NewSTree()
		So(s.isFirst(``), ShouldBeTrue)
		So(s.isFirst(`.key1`), ShouldBeTrue)
		So(s.isFirst(`.key1[0]`), ShouldBeTrue)
		So(s.isFirst(`.key1[1]`), ShouldBeFalse)
		So(s.isFirst(`.key1.key2[0]`), ShouldBeTrue)
		So(s.isFirst(`.key1.key2[1]`), ShouldBeFalse)
		So(s.isFirst(`.key1[0].key2`), ShouldBeTrue)
		So(s.isFirst(`.key1[1].key2`), ShouldBeFalse)
		So(s.isFirst(`.key1.key2[0].key3[0]`), ShouldBeTrue)
		So(s.isFirst(`.key1.key2[1].key3[0]`), ShouldBeFalse)
		So(s.isFirst(`.key1.key2[0].key3[2]`), ShouldBeFalse)
	})
}

type TestStructProto struct {
	IntList  []int `yaml:"intList"`
	TopLevel struct {
		SubStruct struct {
			SubList      []int `yaml:"subList"`
			SubSubStruct struct {
				SubSubKey1 string `yaml:"subSubKey1"`
			} `yaml:"subSubStruct"`
			SubKey1 string `yaml:"subKey1"`
			SubKey2 int    `yaml:"subKey2"`
		} `yaml:"subStruct"`
	} `yaml:"topLevel"`
	IntField   int     `yaml:"intField"`
	FloatField float64 `yaml:"floatField"`
	StrField   string  `yaml:"strField"`
	Product    []struct {
		Sku         string  `yaml:"sku"`
		Quantity    int     `yaml:"quantity"`
		Description string  `yaml:"description"`
		Price       float64 `yaml:"price"`
	} `yaml:"product"`
	StrList   []string `yaml:"strList"`
	Comments  string   `yaml:"comments"`
	BoolField bool     `yaml:"boolField"`
}
