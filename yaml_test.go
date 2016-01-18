package gostree

import (
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

func TestStruct(t *testing.T) {

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

		_, err := findStructElemsPath("", &t, SettingsMap{})
		So(err, ShouldBeNil)
	})

}
