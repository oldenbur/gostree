package gostree

import (
	"strings"
	"testing"

	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {

	testConfig := `
        <seelog type="sync" minlevel="debug">
            <outputs formatid="main"><console/></outputs>
            <formats><format id="main" format="%Date %Time [%LEVEL] %Msg%n"/></formats>
        </seelog>`

	logger, err := log.LoggerFromConfigAsBytes([]byte(testConfig))
	if err != nil {
		panic(err)
	}

	err = log.ReplaceLogger(logger)
	if err != nil {
		panic(err)
	}
}

func TestSTree(t *testing.T) {

	Convey("findStructElems", t, func() {

		type Q struct {
			Q1 int
			Q2 map[int]string
			Q3 string
		}

		type S struct {
			S1  string
			SQ1 Q
			SQ2 Q
		}

		type T struct {
			T1 int
			T2 string
			TS S
			T3 []float64
		}

		t := T{
			T1: 19,
			T2: "goalz",
			TS: S{
				S1:  "many",
				SQ1: Q{Q1: 1, Q2: map[int]string{1: "one", 2: "two"}, Q3: "ONE"},
				SQ2: Q{Q1: 2, Q2: map[int]string{4: "four", 5: "five"}, Q3: "TWO"},
			},
			T3: []float64{1.23, 4.56},
		}

		_, err := findStructElemsPath("", &t, settingsMap{})
		So(err, ShouldBeNil)
	})

	Convey("Json tree access", t, func() {

		s, err := NewSTreeJson(strings.NewReader(`{"key1": "val1", "key2": 1234, "key3": {"key4": true, "key5": -12.34}}`))
		So(err, ShouldBeNil)
		So(s.StrVal("key1"), ShouldEqual, "val1")
		So(s.IntVal("key2"), ShouldEqual, 1234)
		ss := s.STreeVal("key3")
		So(len(ss), ShouldEqual, 2)
		So(s.BoolVal("key3/key4"), ShouldEqual, true)
		So(s.IntVal("key3/key5"), ShouldEqual, -12)

		json, err := s.WriteJson(true)
		So(err, ShouldBeNil)
		log.Debugf("json: %s", string(json))
	})

	Convey("Json Keys", t, func() {

		s, err := NewSTreeJson(strings.NewReader(`{"key1": "val1", "key2": 1234, "key3": {"key4": true, "key5": -12.34}}`))
		So(err, ShouldBeNil)

		keys := s.Keys()
		So(len(keys), ShouldEqual, 3)
		So(keys, ShouldContain, "key1")
		So(keys, ShouldContain, "key2")
		So(keys, ShouldContain, "key3")
		log.Debugf("Json keys: %v", keys)
	})

	Convey("Json KeyStrings", t, func() {

		s, err := NewSTreeJson(strings.NewReader(`{"key1": "val1", "key2": 1234, "key3": {"key4": true, "key5": -12.34}}`))
		So(err, ShouldBeNil)

		keys, err := s.KeyStrings()
		So(err, ShouldBeNil)
		So(len(keys), ShouldEqual, 3)
		So(keys, ShouldContain, "key1")
		So(keys, ShouldContain, "key2")
		So(keys, ShouldContain, "key3")
		log.Debugf("Json key strings: %v", keys)
	})

	Convey("Json array structure\n", t, func() {

		data := `{"a": [{"b": 1, "d": "DDD"}, 19], "c": "bucky"}`
		s, err := NewSTreeJson(strings.NewReader(data))
		So(err, ShouldBeNil)
		log.Debugf("s: %v", s)
		sj, err := s.MarshalJSON()
		So(err, ShouldBeNil)
		log.Debugf("s json: %s", string(sj))
		sl1 := s.SliceVal("a")
		So(len(sl1), ShouldEqual, 2)
		So(s.StrVal("a[0]/d"), ShouldEqual, "DDD")
		So(s.IntVal("a[1]"), ShouldEqual, 19)
		st1 := s.STreeVal("a[0]")
		So(st1.IntVal("b"), ShouldEqual, 1)
		st1j, err := st1.MarshalJSON()
		So(err, ShouldBeNil)
		log.Debugf("st1j: %s", string(st1j))
	})

	Convey("Json array structure\n", t, func() {

		data := `
---
product:
    - sku         : BL394D
      quantity    : 4
      description : Basketball
      price       : 450.00
    - sku         : BL4438H
      quantity    : 1
      description : Super Hoop
      price       : 2392.00
tax  : 251.42
total: 4443.52
comments: >
    Late afternoon is best.
    Backup contact is Nancy
    Billsmer @ 338-4338.
`
		s, err := NewSTreeYaml(strings.NewReader(data))
		So(err, ShouldBeNil)
		log.Debugf("s: %v", s)

	})
}