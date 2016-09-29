package stree

import (
	"strings"
	"testing"

	"fmt"
	. "github.com/smartystreets/goconvey/convey"
)

func init() { InitTestLogger() }

func TestSTreeCmp(t *testing.T) {

	Convey("FieldPaths", t, func() {

		json1 := `{
			"key1": "val1",
			"key2": 1234,
			"key3": {
				"key4": true,
				"key5": -12.34,
				"key6": {
					"key7": false
				},
				"key8": "456"
			},
			"key9": true,
			"key10": true,
			"key11": 4.57,
			"key12": "val12",
			"key13": 986
		}`

		s1, err := NewSTreeJson(strings.NewReader(json1))
		So(err, ShouldBeNil)

		json2 := `{
			"key1": "val1",
			"key2": 1234,
			"key3": {
				"key4": true,
				"key5": -12.34,
				"key6": {
					"key7": true,
					"key9": true
				},
				"key8": 456
			},
			"key10": false,
			"key11": 4.56,
			"key12": "val13",
			"key13": 987
		}`

		s2, err := NewSTreeJson(strings.NewReader(json2))
		So(err, ShouldBeNil)

		cmp, err := s1.CompareTo(s2)
		So(err, ShouldBeNil)
		So(len(cmp), ShouldEqual, 12)

		checkComparison(cmp, ".key1", COMP_NO_DIFFERENCE)
		checkComparison(cmp, ".key2", COMP_NO_DIFFERENCE)
		checkComparison(cmp, ".key3.key4", COMP_NO_DIFFERENCE)
		checkComparison(cmp, ".key3.key5", COMP_NO_DIFFERENCE)
		checkComparison(cmp, ".key3.key6.key7", COMP_VALUES_DIFFER)
		checkComparison(cmp, ".key3.key6.key9", COMP_SUBJECT_LACKS)
		checkComparison(cmp, ".key9", COMP_OBJECT_LACKS)
		checkComparison(cmp, ".key3.key8", COMP_TYPES_DIFFER)
		checkComparison(cmp, ".key10", COMP_VALUES_DIFFER)
		checkComparison(cmp, ".key11", COMP_VALUES_DIFFER)
		checkComparison(cmp, ".key12", COMP_VALUES_DIFFER)
		checkComparison(cmp, ".key13", COMP_VALUES_DIFFER)
	})

}

func checkComparison(cmp ComparisonResult, key string, res FieldComparisonResult) {
	So(cmp, ShouldContainKey, key)
	chk, _ := cmp[key]
	So(fmt.Sprintf("%s %s", key, chk), ShouldEqual, fmt.Sprintf("%s %s", key, res))
}
