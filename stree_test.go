package gostree

import (
	"strings"
	"testing"

	"fmt"
	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
)

func init() { configureTestLogger() }

// ConfigureTestLogger configures the global logger to print to console only
func configureTestLogger() {

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

		s, err := NewSTreeJson(strings.NewReader(`{"key1": "val1", "key.2": 1234, "key3": {"key4": true, "key5": -12.34}}`))
		So(err, ShouldBeNil)
		v1, err := s.StrVal(".key1")
		So(err, ShouldBeNil)
		So(v1, ShouldEqual, "val1")
		v2, err := s.IntVal(`.key\.2`)
		So(err, ShouldBeNil)
		So(v2, ShouldEqual, 1234)
		ss, err := s.STreeVal(".key3")
		So(err, ShouldBeNil)
		So(len(ss), ShouldEqual, 2)
		v4, err := s.BoolVal(".key3.key4")
		So(err, ShouldBeNil)
		So(v4, ShouldEqual, true)
		v5, err := s.FloatVal(".key3.key5")
		So(err, ShouldBeNil)
		So(v5, ShouldEqual, -12.34)

		_, err = s.WriteJson(true)
		So(err, ShouldBeNil)
		//		log.Debugf("json: %s", string(json))
	})

	Convey("Json Keys", t, func() {

		s, err := NewSTreeJson(strings.NewReader(`{"key1": "val1", "key2": 1234, "key3": {"key4": true, "key5": -12.34}}`))
		So(err, ShouldBeNil)

		keys := s.Keys()
		So(len(keys), ShouldEqual, 3)
		So(keys, ShouldContain, "key1")
		So(keys, ShouldContain, "key2")
		So(keys, ShouldContain, "key3")
		//		log.Debugf("Json keys: %v", keys)
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
		//		log.Debugf("Json key strings: %v", keys)
	})

	Convey("Json array structure\n", t, func() {

		data := `{"a": [{"b": 1, "d": "DDD"}, 19], "c": "bucky"}`
		s, err := NewSTreeJson(strings.NewReader(data))
		So(err, ShouldBeNil)
		log.Debugf("s: %v", s)
		sj, err := s.WriteJson(true)
		So(err, ShouldBeNil)
		log.Debugf("s json: %s", string(sj))
		log.Debugf("s: %v", s)
		sl1, err := s.SliceVal(".a")
		So(err, ShouldBeNil)
		So(len(sl1), ShouldEqual, 2)
		vd, err := s.StrVal(".a[0].d")
		So(err, ShouldBeNil)
		So(vd, ShouldEqual, "DDD")
		a1v, err := s.IntVal(".a[1]")
		So(err, ShouldBeNil)
		So(a1v, ShouldEqual, 19)
		st1, err := s.STreeVal(".a[0]")
		So(err, ShouldBeNil)
		st1v, err := st1.IntVal(".b")
		So(err, ShouldBeNil)
		So(st1v, ShouldEqual, 1)
		st1j, err := st1.WriteJson(true)
		So(err, ShouldBeNil)
		log.Debugf("st1j: %s", string(st1j))
	})

	var yamlData string = `
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

	Convey("NewSTreeYaml\n", t, func() {

		s, err := NewSTreeYaml(strings.NewReader(yamlData))
		So(err, ShouldBeNil)
		log.Debugf("s: %v", s)

		_, err = s.WriteYaml()
		So(err, ShouldBeNil)
	})

	Convey("b-tree in yaml\n", t, func() {

		yaml := `
---
L1:
  L2.1:
    L3.1.1:
    L3.1.2:
  L2.2:
    L3.2.1:
`

		s, err := NewSTreeYaml(strings.NewReader(yaml))
		So(err, ShouldBeNil)
		log.Debugf("s: %v", s)

		keys, _ := s.KeyStrings()
		n := newNode(keys[0], s.STreeValMust(AsPath(keys[0])))
		log.Debugf("n:\n%v", n)
	})

	Convey("NewSTreeCopy\n", t, func() {
		s, err := NewSTreeJson(strings.NewReader(`{"key1": "val1", "key2": 1234, "key3": {"key4": true, "key5": -12.34}}`))
		So(err, ShouldBeNil)

		t, err := NewSTreeCopy(s)
		So(err, ShouldBeNil)

		r, err := s.CompareTo(t)
		So(err, ShouldBeNil)
		for _, rr := range r {
			So(rr, ShouldEqual, COMP_NO_DIFFERENCE)
		}

		diffKey := ".key3.key4"
		t, err = t.SetVal(diffKey, false)
		So(err, ShouldBeNil)
		r, err = s.CompareTo(t)
		So(err, ShouldBeNil)
		So(r[diffKey], ShouldEqual, COMP_VALUES_DIFFER)
	})
}

type node struct {
	data string
	l, r *node
}

func newNode(data string, t interface{}) *node {
	log.Debugf("newNode(%s, %v)", data, t)
	n := &node{data: data}
	if t == nil {
		return n
	}
	keys, _ := t.(STree).KeyStrings()
	if len(keys) > 0 {
		lChild, _ := t.(STree).STreeVal(AsPath(keys[0]))
		n.l = newNode(keys[0], lChild)
	}
	if len(keys) > 1 {
		rChild, _ := t.(STree).STreeVal(AsPath(keys[1]))
		n.r = newNode(keys[1], rChild)
	}
	return n
}

func (n *node) String() string {
	return n.string("")
}

func (n *node) string(i string) string {
	s := fmt.Sprintf("%s%s\n", i, n.data)
	if n.l != nil {
		s += n.l.string(i + "  ")
	}
	if n.r != nil {
		s += n.r.string(i + "  ")
	}
	return s
}
